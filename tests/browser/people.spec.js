const { test, expect } = require('./fixtures');
const { expectNightTheme } = require('./night-assertions');

async function expectPeopleOrder(page, names) {
  const links = page.getByRole('list', { name: "People's wishlists" }).getByRole('link');
  await expect(links).toHaveCount(names.length);
  for (const [index, name] of names.entries()) await expect(links.nth(index)).toHaveAccessibleName(name);
}

// @test WLS-001-AC1 WLS-001-AC2 WLS-001-AC6 WLS-001-AC8
// @test WLS-003-AC1 WLS-003-AC2 WLS-003-AC5 WLS-003-AC6 CON-003-AC1 CON-003-AC5
test('parents create people and family members browse their empty current-year lists', async ({ page, context }) => {
  const longName = '𐐀'.repeat(50);
  await page.goto('/');
  await page.getByLabel('Household name').fill('Family');
  await page.getByLabel('Household timezone').fill('Europe/Zurich');
  await page.getByRole('button', { name: 'Create access links' }).click();
  const family = await page.getByLabel('Family link', { exact: true }).inputValue();
  await page.getByRole('button', { name: 'Finish setup' }).click();
  const parent = page.url();
  await expect(page.getByText('No wishlists yet.', { exact: true })).toBeVisible();
  for (const name of ['  Zoë  ', 'Anna', 'Camille']) {
    await page.getByLabel('Person name', { exact: true }).fill(name);
    await page.getByRole('button', { name: 'Add person', exact: true }).click();
    await expect(page).toHaveURL(parent);
    await expect(page.getByRole('list', { name: "People's wishlists" }).getByRole('link', { name: name.trim(), exact: true })).toBeVisible();
  }
  const people = page.getByRole('list', { name: "People's wishlists" });
  await expectPeopleOrder(page, ['Zoë', 'Anna', 'Camille']);
  await page.setViewportSize({ width: 390, height: 844 });
  await expectNightTheme(page);
  await page.screenshot({ path: test.info().outputPath('people-mobile.png'), fullPage: true });
  await page.setViewportSize({ width: 1440, height: 1000 });
  await expectNightTheme(page);
  await page.screenshot({ path: test.info().outputPath('people-desktop.png'), fullPage: true });
  await people.getByRole('link', { name: 'Zoë', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'Zoë’s wishlist' })).toBeVisible();
  await expect(page.getByText('No wishes yet.', { exact: true })).toBeVisible();
  await page.reload();
  await expect(page.getByRole('heading', { name: 'Zoë’s wishlist' })).toBeVisible();
  await expectNightTheme(page);
  await page.getByRole('link', { name: 'All wishlists' }).click();
  await page.getByLabel('Person name', { exact: true }).fill('zoë');
  await page.getByRole('button', { name: 'Add person', exact: true }).click();
  await expect(page.getByRole('alert')).toContainText('already in use');
  await expectPeopleOrder(page, ['Zoë', 'Anna', 'Camille']);
  await page.getByLabel('Person name', { exact: true }).fill('Family!');
  await page.getByRole('button', { name: 'Add person', exact: true }).click();
  await expect(page.getByRole('alert')).toContainText('3–50 Unicode letters and digits');
  await page.getByLabel('Person name', { exact: true }).fill(longName);
  await expect(page.getByLabel('Person name', { exact: true })).toHaveValue(longName);
  await page.getByRole('button', { name: 'Add person', exact: true }).click();
  await expect(page).toHaveURL(parent);
  await page.setViewportSize({ width: 320, height: 740 });
  await expectNightTheme(page);
  const familyPage = await context.newPage();
  await familyPage.goto(family);
  await expect(familyPage.getByLabel('Person name', { exact: true })).toHaveCount(0);
  await expect(familyPage.getByRole('button', { name: 'Add person', exact: true })).toHaveCount(0);
  await expectPeopleOrder(familyPage, ['Zoë', 'Anna', 'Camille', longName]);
  await familyPage.getByRole('link', { name: 'Anna', exact: true }).click();
  await expect(familyPage.getByRole('heading', { name: 'Anna’s wishlist' })).toBeVisible();
  await expect(familyPage.getByText('No wishes yet.', { exact: true })).toBeVisible();
  await expectNightTheme(familyPage);
  await familyPage.close();
});
