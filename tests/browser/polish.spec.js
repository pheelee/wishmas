const { test, expect } = require('./fixtures');
const { expectNightTheme } = require('./night-assertions');
const { readFileSync } = require('node:fs');
const { join } = require('node:path');

// @test CON-003-AC3 CON-003-AC4 CON-003-AC6 CON-003-AC7 CON-003-AC8 CON-003-AC9 ACC-004-AC1 ACC-004-AC2 ACC-004-AC3 ACC-004-AC4
test('setup previews household decoration sliders and persists them for family browsers', async ({ page, context, appDataDir }) => {
 await context.grantPermissions(['clipboard-read', 'clipboard-write']);
 await page.goto('/');
 await expect(page.getByRole('button', { name: /snow/i })).toHaveCount(0);
 const browserZone = await page.evaluate(() => Intl.DateTimeFormat().resolvedOptions().timeZone);
 await expect(page.getByLabel('Household timezone')).toHaveValue(browserZone);
 await expect(page.getByLabel('Bulb size')).toHaveValue('120');
 await expect(page.getByLabel('Snow size')).toHaveValue('100');
 await expect(page.getByLabel('Snow amount')).toHaveValue('16');
 const bulb = page.locator('.bulb').first();
 expect(await bulb.evaluate(e => parseFloat(getComputedStyle(e).width))).toBeCloseTo(7.2, 1);
 const initial = await bulb.evaluate(e => e.getBoundingClientRect().width);
 const count = await page.locator('.bulb').count();
 for (const [label, value] of [['Bulb size','200'], ['Snow size','200'], ['Snow amount','48']]) {
  await page.getByLabel(label).fill(value); await page.getByLabel(label).dispatchEvent('input');
 }
 expect(await page.locator('.bulb').count()).toBeLessThan(count);
 expect(await bulb.evaluate(e => e.getBoundingClientRect().width)).toBeGreaterThan(initial);
 await expect(page.locator('.snowflake')).toHaveCount(48);
 expect(await page.locator('.snowflake').first().evaluate(e => parseFloat(getComputedStyle(e).width))).toBe(2);
 await page.getByLabel('Household timezone').fill('Europe/Zurich');
 await page.getByLabel('Household name').fill('Bad!');
 await page.getByRole('button', { name:'Create access links' }).click();
 await expect(page.getByRole('alert')).toBeVisible();
 await expect(page.getByLabel('Household timezone')).toHaveValue('Europe/Zurich');
 await expect(page.getByLabel('Bulb size')).toHaveValue('200');
 await expect(page.locator('.snowflake')).toHaveCount(48);
 await page.getByLabel('Household name').fill('Family');
 await page.getByRole('button', { name:'Create access links' }).click();
 const family = await page.getByLabel('Family link', { exact:true }).inputValue();
 const copy = page.getByRole('button', { name:'Copy family link', exact:true });
 const inputRect = await page.getByLabel('Family link',{exact:true}).boundingBox();
 const copyRect = await copy.boundingBox();
 expect(Math.abs(inputRect.y + inputRect.height/2 - copyRect.y - copyRect.height/2)).toBeLessThan(2);
 await copy.click(); await expect(copy).toHaveAttribute('data-copied','true');
 expect(await page.evaluate(() => navigator.clipboard.readText())).toBe(family);
 await expect(copy).toHaveAttribute('data-copied','false', { timeout:4500 });
 await page.getByRole('button', { name:'Finish setup' }).click();
 const parent = page.url();
 await page.getByRole('button',{name:'Copy parent link',exact:true}).focus();
 await page.keyboard.press('Enter');
 await expect(page.getByRole('button',{name:'Copy parent link',exact:true})).toHaveAttribute('data-copied','true');
 const data = readFileSync(join(appDataDir,'households','Family','household.yaml'),'utf8');
 expect(data).toContain('bulb_size: 200'); expect(data).toContain('snow_amount: 48');
 const guest = await context.newPage(); await guest.goto(family);
 await expect(guest.locator('.snowflake')).toHaveCount(48);
 await expect(guest.locator('body')).toHaveAttribute('data-bulb-size','200');
 await page.setViewportSize({width:1440,height:1000});
 expect(await page.locator('main').evaluate(e => e.getBoundingClientRect().width)).toBe(1080);
 await page.getByRole('button',{name:'Regenerate family link',exact:true}).click();
 await expect(page.getByRole('dialog')).toBeVisible();
 await page.getByRole('button',{name:'Cancel',exact:true}).click();
 await expect(page).toHaveURL(parent);
 await expect(page.getByLabel('Family link', {exact:true})).toHaveValue(family);
 await page.setViewportSize({width:390,height:844}); await expectNightTheme(page);
 await page.screenshot({path:test.info().outputPath('polish-access-mobile.png'),fullPage:true});
});

// @test WSH-004-AC1 WSH-004-AC2 WSH-004-AC3 WSH-004-AC4 WSH-004-AC5 WLS-005-AC1 WLS-005-AC2 CON-003-AC10
test('wish cards use icon actions and accessible modal forms without changing cancelled data', async ({page,appDataDir}) => {
 await page.goto('/'); await page.getByLabel('Household name').fill('Family');
 await page.getByRole('button',{name:'Create access links'}).click();
 await page.getByRole('button',{name:'Finish setup'}).click();
 await page.getByLabel('Person name',{exact:true}).fill('Anna'); await page.getByRole('button',{name:'Add person',exact:true}).click();
 await page.getByRole('link',{name:'Anna',exact:true}).click();
 const add=page.getByRole('region',{name:'Add wish',exact:true});
 await add.getByLabel('Title',{exact:true}).fill('Wooden train'); await add.getByRole('button',{name:'Add wish',exact:true}).click();
 const file=join(appDataDir,'households','Family',String(new Date().getUTCFullYear()),'christmas','Anna','wishlist.yaml');
 const before=readFileSync(file,'utf8');
 const edit=page.getByRole('button',{name:'Edit wish: Wooden train',exact:true});
 await expect(edit.locator('svg')).toHaveCount(1);
 await expect(page.locator('.wish-card')).toHaveCount(1);
 await edit.click(); const dialog=page.getByRole('dialog'); await expect(dialog).toBeVisible();
 await dialog.getByLabel('Title',{exact:true}).fill('Electric train');
 await page.keyboard.press('Escape'); await expect(dialog).toHaveCount(0); await expect(edit).toBeFocused();
 expect(readFileSync(file,'utf8')).toBe(before);
 await edit.click(); await page.getByRole('dialog').getByRole('button',{name:'Save wish',exact:true}).click();
 await expect(page.getByRole('heading',{name:'Electric train',exact:true})).toBeVisible();
 await page.getByRole('button',{name:'Delete wish: Electric train',exact:true}).click();
 await expect(page.getByRole('dialog')).toBeVisible(); await page.getByRole('button',{name:'Cancel',exact:true}).click();
 await expect(page.getByRole('heading',{name:'Electric train',exact:true})).toBeVisible();
 await page.getByRole('button',{name:'Rename person',exact:true}).click();
 await expect(page.getByRole('dialog').getByLabel('New name')).toBeVisible(); await page.keyboard.press('Escape');
 await page.getByRole('button',{name:'Archive person',exact:true}).click(); await expect(page.getByRole('dialog')).toBeVisible(); await page.keyboard.press('Escape');
 for (const title of ['Storybook set','Winter puzzle']) {
  await add.getByLabel('Title',{exact:true}).fill(title); await add.getByRole('button',{name:'Add wish',exact:true}).click();
 }
 await page.setViewportSize({width:1440,height:1000}); await page.screenshot({path:test.info().outputPath('polish-wishlist-desktop.png'),fullPage:true});
 await page.setViewportSize({width:390,height:844}); await expectNightTheme(page); await page.screenshot({path:test.info().outputPath('polish-wishlist-mobile.png'),fullPage:true});
});
