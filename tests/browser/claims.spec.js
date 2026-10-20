const { test, expect } = require('./fixtures');
const { expectNightTheme } = require('./night-assertions');
const { readFileSync, writeFileSync, mkdirSync } = require('node:fs');
const { join, dirname } = require('node:path');

const card = (page, title = 'Wooden train') => page.locator('.wish-card').filter({ has: page.getByRole('heading', { name: title, exact: true }) });
async function setupClaims(page) {
 await page.goto('/'); await page.getByLabel('Household name').fill('Family'); await page.getByLabel('Household timezone').fill('Europe/Zurich');
 await page.getByRole('button', { name: 'Create access links' }).click();
 const family = await page.getByLabel('Family link', { exact: true }).inputValue();
 await page.getByRole('button', { name: 'Finish setup' }).click(); const parent = page.url();
 const year = Number((await page.getByText(/Parent access/).textContent()).match(/\d+/)[0]);
 await page.getByLabel('Person name', { exact: true }).fill('Anna'); await page.getByRole('button', { name: 'Add person', exact: true }).click();
 await page.getByRole('link', { name: 'Anna', exact: true }).click();
 for (const title of ['Wooden train', 'Another train']) {
  const form = page.getByRole('region', { name: 'Add wish', exact: true });
  await form.getByLabel('Title', { exact: true }).fill(title); await form.getByRole('button', { name: 'Add wish', exact: true }).click();
 }
 return { family, parent, year, list: page.url() };
}
async function familyList(context, family) {
 const page = await context.newPage(); await page.goto(family); await page.getByRole('link', { name: 'Anna', exact: true }).click(); return page;
}
async function firstClaim(page, name) {
 await page.getByRole('button', { name: 'Claim wish: Wooden train', exact: true }).click();
 const prompt = page.getByRole('region', { name: 'Your claim name', exact: true });
 await prompt.getByLabel('Claim name', { exact: true }).fill(name); await prompt.getByRole('button', { name: 'Save name and claim', exact: true }).click();
}

// @test CLM-001-AC1 CLM-001-AC2 CLM-001-AC3 CLM-001-AC4 CLM-001-AC5 CLM-001-AC6 CLM-001-AC7
// @test CLM-002-AC1 CLM-002-AC2 CLM-002-AC3 CLM-002-AC5 CLM-002-AC6 CLM-003-AC2 CLM-003-AC3 CLM-003-AC4 CLM-003-AC5 CLM-003-AC6
// @test CLM-004-AC1 CLM-004-AC2 CLM-004-AC3 CLM-004-AC4 CLM-004-AC5 CLM-004-AC7 CON-003-AC1 CON-003-AC5
test('browser names, multiple claimants, future-only renames, and ownership work across independent browsers', async ({ page, browser, appDataDir }) => {
 test.setTimeout(60000);
 const { family, year } = await setupClaims(page);
 const firstContext = await browser.newContext(), secondContext = await browser.newContext(), sameNameContext = await browser.newContext();
 try {
  const auntie = await familyList(firstContext, family);
  await expect(auntie.getByRole('region', { name: 'Your claim name', exact: true })).toHaveCount(0);
  await auntie.getByRole('button', { name: 'Claim wish: Wooden train', exact: true }).click();
  const prompt = auntie.getByRole('region', { name: 'Your claim name', exact: true });
  await prompt.getByLabel('Claim name', { exact: true }).fill('Al'); await prompt.getByRole('button', { name: 'Save name and claim', exact: true }).click();
  await expect(auntie.getByRole('alert')).toContainText('3–50');
  await prompt.getByLabel('Claim name', { exact: true }).fill(' Auntie '); await prompt.getByRole('button', { name: 'Save name and claim', exact: true }).click();
  await expect(card(auntie).getByText('Claimed by you', { exact: true })).toBeVisible();
  await expect(auntie.getByRole('button', { name: 'Remove my claim: Wooden train', exact: true })).toHaveAttribute('aria-pressed', 'true');
  await auntie.reload(); await expect(auntie.getByText('Claiming as Auntie', { exact: true })).toBeVisible();
  const stored = (await firstContext.cookies()).find(cookie => cookie.name === 'wishmas_claim_name');
  expect(stored.httpOnly).toBe(true); expect(stored.expires).toBeGreaterThan(Date.now() / 1000);
  const uncle = await familyList(secondContext, family); await firstClaim(uncle, 'Uncle');
  await expect(card(auntie).getByText('Claimed · 1 claimant', { exact: true })).toBeVisible();
  await auntie.reload(); await expect(card(auntie).getByText('Claimed · 2 claimants', { exact: true })).toBeVisible();
  expect(await auntie.content()).not.toContain('Uncle');
  await page.reload(); await expect(card(page).locator('.claimant-name')).toHaveText(['Auntie', 'Uncle']);
  const sameName = await familyList(sameNameContext, family); await firstClaim(sameName, 'auntie');
  await expect(sameName.getByRole('alert')).toContainText('already claimed by you');
  await expect(card(sameName).getByText('Claimed · 2 claimants', { exact: true })).toBeVisible();
  await expect(card(sameName).getByText('Claimed by you', { exact: true })).toBeVisible();
  const file = join(appDataDir, 'households', 'Family', String(year), 'christmas', 'Anna', 'wishlist.yaml');
  const beforeRename = readFileSync(file, 'utf8');
  await auntie.getByRole('button', { name: 'Change claim name', exact: true }).click();
  await auntie.getByRole('dialog').getByLabel('Claim name', { exact: true }).fill('Bad!');
  await auntie.getByRole('button', { name: 'Save claim name', exact: true }).click();
  await expect(auntie.getByRole('dialog').getByRole('alert')).toContainText('3–50');
  await auntie.getByRole('dialog').getByLabel('Claim name', { exact: true }).fill(' Cousin ');
  await auntie.getByRole('button', { name: 'Save claim name', exact: true }).click();
  expect(readFileSync(file, 'utf8')).toBe(beforeRename);
  await expect(card(auntie).getByText('Claimed by you', { exact: true })).toHaveCount(0);
  await expect(auntie.getByText('Claiming as Cousin', { exact: true })).toBeVisible();
  expect(await auntie.content()).not.toContain('Auntie');
  await auntie.getByRole('button', { name: 'Claim wish: Wooden train', exact: true }).click();
  await expect(card(auntie).getByText('Claimed · 3 claimants', { exact: true })).toBeVisible();
  await expect(auntie.getByRole('region', { name: 'Your claim name', exact: true })).toHaveCount(0);
  await page.reload(); await expect(card(page).locator('.claimant-name')).toHaveText(['Auntie', 'Uncle', 'Cousin']);
  await auntie.setViewportSize({ width: 390, height: 844 }); await expectNightTheme(auntie);
  await auntie.screenshot({ path: test.info().outputPath('claim-toggle-family-mobile.png'), fullPage: true });
  await auntie.getByRole('button', { name: 'Remove my claim: Wooden train', exact: true }).click();
  await expect(card(auntie).getByText('Claimed · 2 claimants', { exact: true })).toBeVisible();
  await page.reload(); await page.getByRole('button', { name: 'Remove claim by Uncle', exact: true }).click();
  await expect(card(page).locator('.claimant-name')).toHaveText(['Auntie']);
  await sameName.reload(); await sameName.getByRole('button', { name: 'Remove my claim: Wooden train', exact: true }).click();
  await expect(card(sameName).getByText('Unclaimed · 0 claimants', { exact: true })).toBeVisible();
  await page.reload(); await expect(card(page).getByText('Unclaimed · 0 claimants', { exact: true })).toBeVisible();
 } finally { await firstContext.close(); await secondContext.close(); await sameNameContext.close(); }
});

// @test CLM-001-AC6 CLM-002-AC1 CLM-002-AC3 CLM-002-AC4 CLM-002-AC5 CLM-003-AC6 CLM-004-AC6
test('parent toggles, stale duplicate refresh, Unicode names, and historical read-only claims', async ({ page, context, appDataDir }) => {
 const { family, year, list } = await setupClaims(page);
 const longName = '𐐀'.repeat(50);
 await firstClaim(page, longName);
 await expect(card(page).getByText('Claimed by you', { exact: true })).toBeVisible();
 await page.setViewportSize({ width: 390, height: 844 }); await expectNightTheme(page);
 await page.getByRole('button', { name: 'Remove my claim: Wooden train', exact: true }).click();
 const a = await familyList(context, family), b = await familyList(context, family);
 await a.getByRole('button', { name: 'Claim wish: Wooden train', exact: true }).click();
 await b.getByRole('button', { name: 'Claim wish: Wooden train', exact: true }).click();
 await expect(b.getByRole('alert')).toContainText('refreshed');
 await expect(card(b).getByText('Claimed · 1 claimant', { exact: true })).toBeVisible();
 await expect(card(b).getByText('Claimed by you', { exact: true })).toBeVisible();
 const currentFile = join(appDataDir, 'households', 'Family', String(year), 'christmas', 'Anna', 'wishlist.yaml');
 const oldFile = join(appDataDir, 'households', 'Family', String(year - 1), 'christmas', 'Anna', 'wishlist.yaml');
 const saved = readFileSync(currentFile, 'utf8'); mkdirSync(dirname(oldFile), { recursive: true }); writeFileSync(oldFile, saved);
 const history = new URL(list); history.searchParams.set('year', String(year - 1));
 await page.goto(history.toString()); await expect(page.getByText('Read-only history', { exact: true })).toBeVisible();
 await expect(card(page).getByText('Claimed by you', { exact: true })).toBeVisible();
 await expect(page.getByRole('button', { name: /^(Claim wish:|Remove my claim:|Remove claim by)/ })).toHaveCount(0);
 await page.getByRole('button', { name: 'Change claim name', exact: true }).click();
 await page.getByRole('dialog').getByLabel('Claim name', { exact: true }).fill('NewName');
 await page.getByRole('button', { name: 'Save claim name', exact: true }).click();
 await expect(page).toHaveURL(history.toString());
 expect(readFileSync(oldFile, 'utf8')).toBe(saved); expect(readFileSync(currentFile, 'utf8')).toBe(saved);
 await expect(card(page).getByText('Claimed by you', { exact: true })).toHaveCount(0);
 const restricted = new URL(b.url()); restricted.searchParams.set('year', String(year - 1)); expect((await b.goto(restricted.toString())).status()).toBe(404);
});
