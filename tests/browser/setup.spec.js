const { test, expect } = require('./fixtures');

// @test ACC-001-AC1 ACC-001-AC2 ACC-001-AC3 ACC-001-AC4 ACC-001-AC6 ACC-001-AC7 ACC-001-AC8 ACC-001-AC9
// @test ACC-002-AC1 ACC-002-AC5
// @test ACC-003-AC1 ACC-003-AC2 ACC-003-AC3 ACC-003-AC4 ACC-003-AC5 ACC-003-AC6
test('setup, persistent access, independent regeneration, and stale-page revocation', async ({ page, context }) => {
  await context.grantPermissions(['clipboard-read', 'clipboard-write']);
  await page.goto('/');
  await expect(page.getByRole('heading', { name: 'Make room for wishes.' })).toBeVisible();
  await page.screenshot({ path: test.info().outputPath('setup-desktop.png'), fullPage: true });
  await page.setViewportSize({ width: 390, height: 844 });
  await page.screenshot({ path: test.info().outputPath('setup-mobile.png'), fullPage: true });
  await page.getByLabel('Household name').fill('Noël2026');
  await page.getByLabel('Household timezone').fill('Europe/Zurich');
  await page.getByRole('button', { name: 'Create access links' }).click();
  await expect(page.getByRole('heading', { name: 'A Christmas for Noël2026.' })).toBeVisible();
  const parent = await page.getByLabel('Parent link', { exact: true }).inputValue();
  const family = await page.getByLabel('Family link', { exact: true }).inputValue();
  expect(parent).not.toBe(family);
  expect(parent).toMatch(/^http:\/\/127\.0\.0\.1:\d+\/\?code=/);
  for (const [label, value] of [['Copy parent link', parent], ['Copy family link', family]]) {
    await page.getByRole('button', { name: label }).click();
    await expect(page.getByRole('status')).toHaveText('Link copied.');
    expect(await page.evaluate(() => navigator.clipboard.readText())).toBe(value);
  }
  await page.getByRole('button', { name: 'Finish setup' }).click();
  await expect(page).toHaveURL(parent);
  await expect(page.getByText(/Parent access/)).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Christmas at Noël2026.' })).toBeVisible();
  await expect(page.getByLabel('Household timezone')).toHaveCount(0);
  await page.reload();
  await expect(page.getByText(/Parent access/)).toBeVisible();
  await page.goto(family);
  await expect(page.getByText(/Family access/)).toBeVisible();
  const missingCode = await page.goto('/');
  expect(missingCode.status()).toBe(404);
  const closedWizard = await page.goto('/setup');
  expect(closedWizard.status()).toBe(404);

  await page.goto(parent);
  const staleParentPage = await context.newPage();
  await staleParentPage.goto(parent);
  const familyPage = await context.newPage();
  await familyPage.goto(family);
  await expect(familyPage.getByRole('button', { name: /Regenerate/ })).toHaveCount(0);
  await expect(familyPage.getByLabel('Parent link', { exact: true })).toHaveCount(0);

  await page.getByRole('button', { name: 'Regenerate family link', exact: true }).click();
  await expect(page).toHaveURL(parent);
  await expect(page.getByLabel('Family link', { exact: true })).not.toHaveValue(family);
  const newFamily = await page.getByLabel('Family link', { exact: true }).inputValue();
  await expect(page.getByLabel('Parent link', { exact: true })).toHaveValue(parent);
  const revokedFamily = await familyPage.reload();
  expect(revokedFamily.status()).toBe(404);
  await familyPage.goto(newFamily);
  await expect(familyPage.getByText(/Family access/)).toBeVisible();

  await page.getByRole('button', { name: 'Regenerate parent link', exact: true }).click();
  await expect(page).not.toHaveURL(parent);
  await expect(page.getByText(/Parent access/)).toBeVisible();
  const newParent = await page.getByLabel('Parent link', { exact: true }).inputValue();
  await expect(page).toHaveURL(newParent);
  await expect(page.getByLabel('Family link', { exact: true })).toHaveValue(newFamily);
  await page.getByRole('button', { name: 'Copy parent link', exact: true }).click();
  await expect.poll(() => page.evaluate(() => navigator.clipboard.readText())).toBe(newParent);
  await page.screenshot({ path: test.info().outputPath('parent-access-mobile.png'), fullPage: true });

  // Use the stale page's form directly, without reloading it first.
  const revokedAction = staleParentPage.waitForResponse(response => response.request().method() === 'POST');
  await staleParentPage.getByRole('button', { name: 'Regenerate family link', exact: true }).click();
  expect((await revokedAction).status()).toBe(404);
  await expect(staleParentPage.getByText('404 page not found')).toBeVisible();
  await page.reload();
  await expect(page.getByLabel('Family link', { exact: true })).toHaveValue(newFamily);
  const oldParent = await staleParentPage.goto(parent);
  expect(oldParent.status()).toBe(404);
  await familyPage.reload();
  await expect(familyPage.getByText(/Family access/)).toBeVisible();
  await staleParentPage.close();
  await familyPage.close();
});
