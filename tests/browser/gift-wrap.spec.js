const { test, expect } = require('./fixtures');
const { expectNightTheme } = require('./night-assertions');
const { readFileSync, writeFileSync, mkdirSync } = require('node:fs');
const { join, dirname } = require('node:path');

// @test WSH-005-AC1 WSH-005-AC2 WSH-005-AC3 WSH-005-AC4 WSH-005-AC5 WSH-005-AC6 WSH-005-AC7
test('Gift-wrap cards show badges, price accents, optional details and responsive action footers', async ({ page, context, appDataDir }) => {
 await page.goto('/'); await page.getByLabel('Household name').fill('Family'); await page.getByLabel('Household timezone').fill('Europe/Zurich');
 await page.getByRole('button', { name: 'Create access links' }).click();
 const family = await page.getByLabel('Family link', { exact: true }).inputValue();
 await page.getByRole('button', { name: 'Finish setup' }).click();
 const year = Number((await page.getByText(/Parent access/).textContent()).match(/\d+/)[0]);
 await page.getByLabel('Person name', { exact: true }).fill('Anna'); await page.getByRole('button', { name: 'Add person', exact: true }).click();
 const file = join(appDataDir, 'households', 'Family', String(year), 'christmas', 'Anna', 'wishlist.yaml');
 const sample = 'wishes:\n  - title: Wooden train set\n    note: Small wooden tracks and magnetic carriages. Something we can build together.\n    price: "49.90"\n    links: [https://toys.example/train, https://shop.example/train]\n  - title: Illustrated storybook set\n    note: A few new stories for the quiet evenings before Christmas.\n    price: "24.50"\n    links: [https://books.example/stories]\n    claims: [{name: Auntie}, {name: Uncle}]\n  - title: Winter puzzle\n    claims: [{name: Grandma}]\n';
 writeFileSync(file, sample);
 await context.addCookies([{ name: 'wishmas_claim_name', value: 'Auntie', url: family }]);
 await page.getByRole('link', { name: 'Anna', exact: true }).click();
 await expect(page.locator('.wish-gift-header')).toHaveCount(3);
 await expect(page.locator('.wish-gift-icon svg')).toHaveCount(3);
 await expect(page.locator('.wish-badge')).toHaveCount(4);
 const cards = page.locator('.wish-card');
 const colors = await cards.evaluateAll(elements => elements.map(e => getComputedStyle(e).getPropertyValue('--wish-accent').trim()));
 expect(new Set(colors).size).toBe(3);
 async function expectStatusColors(viewer) {
  const list=viewer.locator('.wish-card');
  for (const [index,status,color] of [[0,'yellow','rgb(242, 209, 140)'],[1,'green','rgb(134, 220, 160)'],[2,'red','rgb(255, 158, 154)']]) {
   const card=list.nth(index);
   await expect(card).toHaveClass(new RegExp(`wish-accent-${status}`));
   await expect(card).toHaveCSS('border-left-color',color);
   await expect(card).toHaveCSS('background-color','rgb(16, 29, 48)');
   await expect(card.locator('.wish-gift-icon')).toHaveCSS('color',color);
   if(await card.locator('.wish-price strong').count()) await expect(card.locator('.wish-price strong')).toHaveCSS('color',color);
   const claim=card.locator('.wish-claim-button[aria-pressed="false"]');
   if(await claim.count()) await expect(claim).toHaveCSS('background-color',color);
  }
 }
 await expectStatusColors(page);
 await page.locator('[data-wish-sort]').selectOption('newest');
 await expect(page.locator('.wish-card').first()).toHaveClass(/wish-accent-red/);
 await page.locator('[data-wish-sort]').selectOption('oldest');

 expect(await cards.first().locator('.wish-title').evaluate(e => getComputedStyle(e).fontFamily)).toContain('Georgia');
 await expect(cards.nth(2).locator('.wish-price')).toHaveCount(0);
 await expect(cards.nth(2).locator('.wish-note')).toHaveCount(0);
 await expect(cards.nth(2).locator('.wish-shops')).toHaveCount(0);
 await expect(cards.first().getByRole('link')).toHaveText(['toys.example', 'shop.example']);
 await expect(cards.nth(1).getByText('Claimed by you', { exact: true })).toBeVisible();
 const footer = cards.first().locator('.wish-card-footer');
 const claim = await footer.getByRole('button', { name: 'Claim wish: Wooden train set', exact: true }).boundingBox();
 const edit = await footer.getByRole('button', { name: 'Edit wish: Wooden train set', exact: true }).boundingBox();
 expect(claim.x).toBeLessThan(edit.x); expect(Math.abs(claim.y + claim.height/2 - edit.y - edit.height/2)).toBeLessThan(2);
 // Header contents keep interior spacing and share a vertical center with one or two badges.
 async function expectAlignedHeaders() {
  for (const header of await page.locator('.wish-gift-header').all()) {
   const box = await header.boundingBox();
   const icon = await header.locator('.wish-gift-icon').boundingBox();
   const badges = await header.locator('.wish-badges').boundingBox();
   expect(box.y + box.height - badges.y - badges.height).toBeGreaterThanOrEqual(19);
   expect(Math.abs(icon.y + icon.height/2 - badges.y - badges.height/2)).toBeLessThan(2);
  }
 }
 await page.setViewportSize({ width:1440, height:1000 }); await expectNightTheme(page);
 await expectAlignedHeaders();
 await page.screenshot({ path:test.info().outputPath('gift-wrap-desktop.png'), fullPage:true });
 await page.setViewportSize({ width:390, height:844 }); await expectNightTheme(page);
 await expectAlignedHeaders();
 const first = await cards.nth(0).boundingBox(), second = await cards.nth(1).boundingBox();
 expect(first.x).toBe(second.x); expect(second.y).toBeGreaterThan(first.y+first.height);
 await page.screenshot({ path:test.info().outputPath('gift-wrap-mobile.png'), fullPage:true });
 const guest = await context.newPage(); await guest.goto(family); await guest.getByRole('link', { name:'Anna', exact:true }).click();
 expect(await guest.content()).not.toMatch(/Uncle|Grandma/);
 await expect(guest.getByRole('button', { name:/^(Edit wish:|Delete wish:|Remove claim by)/ })).toHaveCount(0);
 await expectNightTheme(guest);
 await expectStatusColors(guest);
 expect(readFileSync(file,'utf8')).toBe(sample);
 // Historical cards keep their styling but omit all mutation controls.
 const old = join(appDataDir,'households','Family',String(year-1),'christmas','Anna','wishlist.yaml'); mkdirSync(dirname(old),{recursive:true}); writeFileSync(old,sample);
 const history = new URL(page.url()); history.searchParams.set('year',String(year-1)); await page.goto(history.toString());
 await expect(page.locator('.wish-gift-header')).toHaveCount(3);
 await expect(page.locator('.wish-card-footer')).toHaveCount(0);
 await expect(page.getByRole('button',{name:/^(Claim wish:|Remove my claim:|Remove claim by|Edit wish:|Delete wish:)/})).toHaveCount(0);
 expect(readFileSync(old,'utf8')).toBe(sample);
 await expectStatusColors(page);
 // Status updates on page load when claims or completeness change.
 await guest.getByRole('button',{name:'Claim wish: Wooden train set',exact:true}).click();
 await expect(guest.locator('.wish-card').first()).toHaveClass(/wish-accent-green/);
 await guest.getByRole('button',{name:'Remove my claim: Wooden train set',exact:true}).click();
 await expect(guest.locator('.wish-card').first()).toHaveClass(/wish-accent-yellow/);
 const current=new URL(page.url());current.searchParams.set('year',String(year));await page.goto(current.toString());
 await page.getByRole('button',{name:'Edit wish: Winter puzzle',exact:true}).click();
 let modal=page.locator('dialog[open]');
 await modal.getByLabel('Reference price',{exact:true}).fill('0');
 await modal.getByLabel('Shop links',{exact:true}).fill('https://shop.example/puzzle');
 await modal.getByRole('button',{name:'Save wish',exact:true}).click();
 await expect(page.locator('.wish-card').nth(2)).toHaveClass(/wish-accent-green/);
 await page.getByRole('button',{name:'Edit wish: Winter puzzle',exact:true}).click();
 modal=page.locator('dialog[open]');await modal.getByLabel('Reference price',{exact:true}).fill('');
 await modal.getByRole('button',{name:'Save wish',exact:true}).click();
 await expect(page.locator('.wish-card').nth(2)).toHaveClass(/wish-accent-red/);

});
