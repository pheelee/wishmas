const { test, expect } = require('./fixtures');
const { readFileSync, writeFileSync, mkdirSync, existsSync } = require('node:fs');
const { join, dirname } = require('node:path');
const { expectNightTheme } = require('./night-assertions');

// @test WSH-006-AC1 WSH-006-AC2 WSH-006-AC7 WSH-006-AC8 WSH-006-AC9 WSH-006-AC10 WSH-006-AC12 WSH-006-AC13 WSH-006-AC14 WSH-006-AC15
for (const locale of ['en-US', 'de-CH']) test(`Picture shelf, modal, drafts and cleanup in ${locale}`, async ({page, appDataDir}) => {
 await page.goto('/');
 await page.getByLabel('Household name').fill('Family');
 await page.getByLabel('Household timezone').fill('Europe/Zurich');
 await page.getByRole('button',{name:'Create access links'}).click();
 const family = await page.getByLabel('Family link',{exact:true}).inputValue();
 await page.getByRole('button',{name:'Finish setup'}).click();
 const parent = page.url();
 const year = (await page.getByText(/Parent access/).textContent()).match(/\d+/)[0];
 await page.getByLabel('Person name',{exact:true}).fill('Anna');
 await page.getByRole('button',{name:'Add person',exact:true}).click();
 const file = join(appDataDir,'households','Family',year,'christmas','Anna','wishlist.yaml');
 const pictures = join(dirname(file),'pictures'); mkdirSync(pictures,{recursive:true});
 const webp = await page.evaluate(() => {
  const c=document.createElement('canvas');c.width=640;c.height=320;const ctx=c.getContext('2d');
  ctx.fillStyle='#f2d18c';ctx.fillRect(0,0,640,320);ctx.fillStyle='#bd4d53';ctx.fillRect(220,100,200,120);
  return c.toDataURL('image/webp',.8).split(',')[1];
 });
 writeFileSync(join(pictures,'sample.webp'),Buffer.from(webp,'base64'));
 writeFileSync(file,'wishes:\n - title: Picture train\n   note: Text stays outside the picture\n   picture: sample.webp\n   picture_source: https://images.example/train.png\n   price: "12.50"\n   links: [https://shop.example/train]\n - title: Plain wooden train\n - title: Missing picture\n   picture: missing.webp\n   picture_source: https://images.example/missing.png\n');
 if(locale==='de-CH') {
  const config=join(appDataDir,'households','Family','household.yaml');
  writeFileSync(config,readFileSync(config,'utf8').replace('language: en-US','language: de-CH'));
 }
 const de=locale==='de-CH', imageLabel=de?'Bild-URL':'Image URL', editLabel=de?'Wunsch bearbeiten: Picture train':'Edit wish: Picture train';
 await page.getByRole('link',{name:'Anna',exact:true}).click();
 const cards=page.locator('.wish-card'), first=cards.first();
 // @test WSH-006-AC10: inspect actual running animations on both transitions.
 const pictureModal=page.locator('.picture-modal').first();
 const opening=await first.locator('.wish-picture-button').evaluate(button=>{
  button.click();return document.querySelector('.picture-modal').getAnimations().map(a=>a.effect.getTiming().easing);
 });
 expect(opening).toContain('ease-in-out');
 await expect(pictureModal).toBeVisible();
 const closing=await pictureModal.evaluate(modal=>{
  modal.querySelector('[data-close-modal]').click();
  return {open:modal.open,easing:modal.getAnimations().map(a=>a.effect.getTiming().easing)};
 });
 expect(closing.open).toBe(true);expect(closing.easing).toContain('ease-in-out');
 await expect(pictureModal).not.toBeVisible();
 await page.emulateMedia({reducedMotion:'reduce'});
 await first.locator('.wish-picture-button').click();
 expect(await pictureModal.evaluate(modal=>modal.getAnimations().length)).toBe(0);
 await pictureModal.getByRole('button').click();await expect(pictureModal).not.toBeVisible();
 await page.emulateMedia({reducedMotion:'no-preference'});

 await expect(first.locator('.wish-picture')).toBeVisible();
 await expect(cards.nth(1).locator('[data-picture-shelf]')).toHaveCount(0);
 await expect(cards.nth(2).getByText(de?'Bild nicht verfügbar':'Image unavailable',{exact:true})).toBeVisible();
 expect(await first.locator('.wish-picture').evaluate(e=>getComputedStyle(e).objectPosition)).toBe('50% 50%');
 for(const viewport of [{width:1440,height:1000},{width:390,height:844}]) {
  await page.setViewportSize(viewport);await expectNightTheme(page);
  const title=await first.locator('.wish-title').boundingBox(), note=await first.locator('.wish-note').boundingBox(), image=await first.locator('.wish-picture').boundingBox(), details=await first.locator('.wish-details').boundingBox(), footer=await first.locator('.wish-card-footer').boundingBox();
  expect(image.y).toBeGreaterThanOrEqual(note.y+note.height);expect(title.y).toBeLessThan(image.y);expect(details.y+details.height).toBeLessThanOrEqual(footer.y);
  expect(image.x+image.width).toBeLessThanOrEqual(viewport.width);
  await first.locator('.wish-picture-button').click();
  const modal=page.locator('dialog[open]');await expect(modal.locator('img')).toBeVisible();
  expect(await modal.locator('img').evaluate(e=>getComputedStyle(e).objectFit)).toBe('contain');
  await page.screenshot({path:test.info().outputPath(`picture-${locale}-${viewport.width}.png`),fullPage:true});
  await modal.locator('img').click();await expect(modal).toHaveCount(1);
  const bounds=await modal.boundingBox();
  await page.mouse.click(bounds.x+3,bounds.y+3);await expect(modal).toHaveCount(1);
  await page.mouse.click(2,2);await expect(modal).toHaveCount(0);
  await first.locator('.wish-picture-button').click();
  await page.keyboard.press('Escape');await expect(modal).toHaveCount(0);
 }
 await first.getByRole('button',{name:editLabel,exact:true}).click();
 let modal=page.locator('dialog[open]');await expect(modal.getByLabel(imageLabel,{exact:true})).toHaveValue('https://images.example/train.png');
 await modal.getByLabel(imageLabel,{exact:true}).fill('http://127.0.0.1/private.png');
 await modal.getByLabel(de?'Notiz':'Note',{exact:true}).fill('Draft is retained');
 await modal.getByLabel(de?'Titel':'Title',{exact:true}).fill('  Picture train  ');
 await modal.getByLabel(de?'Shop-Links':'Shop links',{exact:true}).fill(' https://shop.example/train \n\nhttps://other.example/train');
 const before=readFileSync(file,'utf8');
 await modal.getByRole('button',{name:de?'Wunsch speichern':'Save wish',exact:true}).click();
 modal=page.locator('dialog[open]');
 await expect(modal.getByLabel(imageLabel,{exact:true})).toHaveValue('http://127.0.0.1/private.png');
 await expect(modal.getByLabel(de?'Notiz':'Note',{exact:true})).toHaveValue('Draft is retained');
 await expect(modal.getByLabel(de?'Titel':'Title',{exact:true})).toHaveValue('  Picture train  ');
 await expect(modal.getByLabel(de?'Shop-Links':'Shop links',{exact:true})).toHaveValue(' https://shop.example/train \n\nhttps://other.example/train');
 await expect(modal.getByRole('alert')).toContainText(de?'öffentlichen Internetadressen':'public internet addresses');
 expect(readFileSync(file,'utf8')).toBe(before);
 // An unchanged URL saves without contacting the intentionally nonexistent source.
 await modal.getByLabel(imageLabel,{exact:true}).fill('https://images.example/train.png');
 await modal.getByRole('button',{name:de?'Korrigierten Wunsch speichern':'Save corrected wish',exact:true}).click();
 await expect(page.locator('dialog[open]')).toHaveCount(0);
 expect(existsSync(join(pictures,'sample.webp'))).toBe(true);
 await page.goto(family);await page.getByRole('link',{name:'Anna',exact:true}).click();
 await expect(page.locator('.wish-picture-button')).toHaveCount(1);
 await expect(page.locator('input[name="picture_url"]')).toHaveCount(0);
 await page.goto(parent);await page.getByRole('link',{name:'Anna',exact:true}).click();
 await page.getByRole('button',{name:editLabel,exact:true}).click();
 modal=page.locator('dialog[open]');await modal.getByLabel(imageLabel,{exact:true}).fill('');
 await modal.getByRole('button',{name:de?'Wunsch speichern':'Save wish',exact:true}).click();
 await expect(page.locator('.wish-card').first().locator('[data-picture-shelf]')).toHaveCount(0);
 expect(existsSync(join(pictures,'sample.webp'))).toBe(false);
});
