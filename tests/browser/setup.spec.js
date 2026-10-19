const { test, expect } = require('@playwright/test');

// @test ACC-001-AC1 ACC-001-AC2 ACC-001-AC3 ACC-001-AC4 ACC-001-AC6 ACC-001-AC7 ACC-001-AC8 ACC-001-AC9
// @test ACC-002-AC1 ACC-002-AC5
test('household setup, copy links, and finish into parent access', async ({ page, context }) => {
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
  expect(parent).toMatch(/^http:\/\/127\.0\.0\.1:18080\/\?code=/);
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
 });
