const { test, expect } = require('./fixtures');
const { expectNightTheme } = require('./night-assertions');
const { readFileSync, writeFileSync, mkdirSync } = require('node:fs');
const { join, dirname } = require('node:path');

async function setupLists(page, appDataDir) {
 await page.goto('/');
 await page.getByLabel('Household name').fill('Family');
 await page.getByLabel('Household timezone').fill('Europe/Zurich');
 await page.getByRole('button', { name: 'Create access links' }).click();
 const family = await page.getByLabel('Family link', { exact: true }).inputValue();
 await page.getByRole('button', { name: 'Finish setup' }).click();
 const parent = page.url();
 const year = Number((await page.getByText(/Parent access/).textContent()).match(/\d+/)[0]);
 for (const name of ['Anna', 'Bobby']) {
  await page.getByLabel('Person name', { exact: true }).fill(name);
  await page.getByRole('button', { name: 'Add person', exact: true }).click();
 }
 const wishes = [
  { title: 'Zebra gift', price: '10' },
  { title: 'éclair gift', price: '2' },
  { title: 'Apple gift', price: '100' },
  { title: 'ECLAIR gift' },
  { title: 'e\u0301clair gift', price: '2.00' },
  { title: 'Zero gift', price: '0' },
  { title: 'Precise larger', price: '9007199254740992.02' },
  { title: 'Precise smaller', price: '9007199254740992.01' },
 ];
 const yaml = 'wishes:\n' + wishes.map(w => `  - title: ${w.title}\n${w.price === undefined ? '' : `    price: "${w.price}"\n`}`).join('');
 const paths = [];
 for (const [name, savedYear] of [['Anna', year], ['Bobby', year], ['Anna', year - 1]]) {
  const file = join(appDataDir, 'households', 'Family', String(savedYear), 'christmas', name, 'wishlist.yaml');
  mkdirSync(dirname(file), { recursive: true }); writeFileSync(file, yaml); paths.push(file);
 }
 await page.getByRole('link', { name: 'Anna', exact: true }).click();
 return { family, parent, year, wishes, yaml, paths };
}

const titles = page => page.locator('.wish-title');

// @test WSH-003-AC1 WSH-003-AC2 WSH-003-AC3 WSH-003-AC4 WSH-003-AC5 WSH-003-AC6 WSH-003-AC7 WSH-003-AC8 CON-003-AC1 CON-003-AC5
test('wish sorting uses numeric prices, stable accent-insensitive ties, and a shared browser preference', async ({ page, context, appDataDir }) => {
 const { family, year, wishes, yaml, paths } = await setupLists(page, appDataDir);
 const oldest = wishes.map(w => w.title);
 const low = [oldest[3], oldest[5], oldest[1], oldest[4], oldest[0], oldest[2], oldest[7], oldest[6]];
 const high = [oldest[6], oldest[7], oldest[2], oldest[0], oldest[1], oldest[4], oldest[3], oldest[5]];
 const alphabetical = [oldest[2], oldest[1], oldest[3], oldest[4], oldest[6], oldest[7], oldest[0], oldest[5]];
 await expect(titles(page)).toHaveText(oldest);
 const sort = page.getByLabel('Sort wishes', { exact: true });
 await expect(sort).toHaveValue('oldest');
 for (const [value, order] of [['newest', [...oldest].reverse()], ['title', alphabetical], ['price-asc', low], ['price-desc', high], ['oldest', oldest]]) {
  await sort.selectOption(value);
  await expect(titles(page)).toHaveText(order);
 }
 await sort.selectOption('title');
 await page.reload(); await expect(sort).toHaveValue('title'); await expect(titles(page)).toHaveText(alphabetical);
 // Stable ties use original added order even after reversing the displayed list.
 await sort.selectOption('newest'); await sort.selectOption('price-asc'); await expect(titles(page)).toHaveText(low);
 await page.getByRole('link', { name: 'All wishlists' }).click();
 await page.getByRole('link', { name: 'Bobby', exact: true }).click();
 await expect(sort).toHaveValue('price-asc'); await expect(titles(page)).toHaveText(low);
 const familyPage = await context.newPage(); await familyPage.goto(family);
 await familyPage.getByRole('link', { name: 'Anna', exact: true }).click();
 await expect(familyPage.getByLabel('Sort wishes')).toHaveValue('price-asc'); await expect(titles(familyPage)).toHaveText(low);
 await familyPage.getByLabel('Sort wishes').selectOption('price-desc'); await expect(titles(familyPage)).toHaveText(high);
 // Historical lists can be sorted without writing historical data.
 const historyURL = new URL(page.url()); historyURL.searchParams.set('person', 'Anna'); historyURL.searchParams.set('year', String(year - 1));
 await page.goto(historyURL.toString());
 await expect(sort).toHaveValue('price-desc'); await expect(titles(page)).toHaveText(high);
 await expect(page.getByText('Read-only history', { exact: true })).toBeVisible();
 await sort.selectOption('oldest'); await expect(titles(page)).toHaveText(oldest);
 await page.setViewportSize({ width: 390, height: 844 }); await expectNightTheme(page);
 await page.screenshot({ path: test.info().outputPath('sorting-mobile.png'), fullPage: true });
 // A new browser has no inherited preference.
 const freshContext = await page.context().browser().newContext();
 try {
  const fresh = await freshContext.newPage(); await fresh.goto(family);
  await fresh.getByRole('link', { name: 'Anna', exact: true }).click();
  await expect(fresh.getByLabel('Sort wishes')).toHaveValue('oldest'); await expect(titles(fresh)).toHaveText(oldest);
 } finally { await freshContext.close(); }
 for (const file of paths) expect(readFileSync(file, 'utf8')).toBe(yaml);
});

// @test WSH-003-AC1 WSH-003-AC2 WSH-003-AC5
test('sorting remains usable with unavailable browser storage or an invalid saved preference', async ({ page, appDataDir }) => {
 const { wishes } = await setupLists(page, appDataDir);
 await page.evaluate(() => localStorage.setItem('wishmas.wishSort', 'invalid-order'));
 await page.reload();
 const sort = page.getByLabel('Sort wishes', { exact: true });
 await expect(sort).toHaveValue('oldest'); await expect(titles(page)).toHaveText(wishes.map(w => w.title));
 await page.addInitScript(() => {
  Storage.prototype.getItem = () => { throw new Error('Storage disabled'); };
  Storage.prototype.setItem = () => { throw new Error('Storage disabled'); };
 });
 await page.reload(); await expect(sort).toHaveValue('oldest');
 await sort.selectOption('newest'); await expect(titles(page)).toHaveText(wishes.map(w => w.title).reverse());
});

// @test WSH-003-AC1 WSH-003-AC2 WSH-002-AC1 WSH-002-AC6
test('sorting keeps edit and delete actions attached to their original wishes', async ({ page, appDataDir }) => {
 const { wishes } = await setupLists(page, appDataDir);
 const sort = page.getByLabel('Sort wishes', { exact: true });
 await sort.selectOption('newest');
 const zebra = page.locator('[data-wish-list] > li').filter({ has: page.getByRole('heading', { name: 'Zebra gift', exact: true }) });
 await zebra.getByRole('button', { name: 'Edit wish: Zebra gift', exact: true }).click();
 await zebra.getByLabel('Title', { exact: true }).fill('Updated zebra');
 await zebra.getByRole('button', { name: 'Save wish', exact: true }).click();
 const updated = wishes.map(w => w.title); updated[0] = 'Updated zebra';
 await expect(sort).toHaveValue('newest'); await expect(titles(page)).toHaveText([...updated].reverse());
 await sort.selectOption('oldest'); await expect(titles(page)).toHaveText(updated);
 await sort.selectOption('title');
 await page.getByRole('button', { name: 'Delete wish: Apple gift', exact: true }).click();
 await expect(page.getByRole('dialog').getByText('Apple gift', { exact: true })).toBeVisible();
 await page.getByRole('button', { name: 'Permanently delete wish', exact: true }).click();
 await sort.selectOption('oldest');
 await expect(titles(page)).toHaveText(updated.filter(title => title !== 'Apple gift'));
});
