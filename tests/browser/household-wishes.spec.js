const {test,expect}=require('./fixtures');
const {expectNightTheme}=require('./night-assertions');
const {readFileSync,writeFileSync,mkdirSync,existsSync}=require('node:fs');
const {join}=require('node:path');
const messages=require('../../internal/app/messages.json');
test.use({ignoreHTTPSErrors:true});
const password='PublicWishFixture1!';
// @test WSH-001 WSH-003 WSH-004 WSH-005 WSH-006 WSH-007 WLS-010 WLS-011
for(const locale of ['en-US','de-CH'])test(`Household wish lifecycle, assignment, summaries and pictures in ${locale}`,async({page,browser,baseURL,appDataDir,appControl})=>{
 test.setTimeout(120000);page.setDefaultTimeout(12000);
 const t=(key,...args)=>{let text=messages[key][locale==='de-CH'?'de':'en'];let i=0;return text.replace(/\{[^}]+\}/g,()=>String(args[i++]));};
 await page.goto('/');await page.locator('[name=household]').fill('Family');await page.locator('[name=language]').selectOption(locale);await page.locator('[name=password]').fill(password);await page.locator('[name=password_confirm]').fill(password);await page.getByRole('button',{name:'Finish setup',exact:true}).click();
 for(const [kind,name] of [['recipient','Child Name'],['recipient','Other Child'],['user','Viewer User'],['user','Private Claimant'],['list','Christmas 2001'],['list','Birthday 2027']]){await page.locator(`#${kind}-name`).fill(name);await page.locator(`form[action="/manage/${kind}/create"] button`).click();}
 const listLinks={};
 for(const list of ['Christmas 2001','Birthday 2027']){
  await page.getByRole('link',{name:list,exact:true}).click();await page.locator('#recipient-select').selectOption({label:'Child Name'});await page.getByRole('button',{name:t('HR50'),exact:true}).click();
  listLinks[list]=await page.locator('[data-access-link]').first().inputValue();
  if(list==='Christmas 2001'){await page.locator('#recipient-select').selectOption({label:'Other Child'});await page.getByRole('button',{name:t('HR50'),exact:true}).click();}
  await page.getByRole('link',{name:t('HR24'),exact:true}).click();
 }
 // Inline title-only creation starts unassigned. Optional values can be added later.
 const create=page.locator('form[action="/wishes/save"]').last();await create.getByLabel(t('W24'),{exact:true}).fill('Future train');await create.getByRole('button',{name:t('W23'),exact:true}).click();
 let card=page.locator('.wish-card').filter({has:page.getByRole('heading',{name:'Future train',exact:true})});await expect(card).toHaveClass(/wish-accent-red/);await expect(card).toContainText(t('HR76'));
 await card.getByRole('button',{name:t('W13','Future train'),exact:true}).click();let modal=page.locator('dialog[open]');await modal.getByLabel(t('W10'),{exact:true}).fill('12,50');await modal.getByRole('button',{name:t('W14'),exact:true}).click();modal=page.locator('dialog[open]');await expect(modal.getByRole('alert')).toHaveText(t('E40'));await expect(modal.getByLabel(t('W10'),{exact:true})).toHaveValue('12,50');
 await modal.getByLabel(t('W10'),{exact:true}).fill('12.50');await modal.getByLabel(t('W26'),{exact:true}).fill('https://first.example/train\nhttps://second.example/train');await modal.getByLabel(t('W28'),{exact:true}).fill('Text stays outside picture');await modal.getByRole('button',{name:t('W22'),exact:true}).click();
 card=page.locator('.wish-card').filter({has:page.getByRole('heading',{name:'Future train',exact:true})});await expect(card).toHaveClass(/wish-accent-yellow/);await expect(card.locator('.wish-shops a')).toHaveText(['first.example','second.example']);
 // Assign from the household pool; move directly between selected recipients.
 const assign=async(label)=>{await card.getByRole('button',{name:t('HR79')+' · Future train',exact:true}).click();const d=page.locator('dialog[open]');await d.locator('[name=assignment]').selectOption({label});await d.getByRole('button',{name:t('HR81'),exact:true}).click();};
 await assign('Child Name · Christmas 2001');
 // Seed an optimized local image and ID-based claim without introducing claim-action implementation.
 const cfg=join(appDataDir,'households/Family/household.yaml');let raw=readFileSync(cfg,'utf8');const id=raw.match(/wishes:\n\s+- id: (\w+)/)[1],userID=[...raw.split('users:\n')[1].split('recipients:')[0].matchAll(/- id: (\w+)/g)][1][1];
 const webp=await page.evaluate(()=>{const c=document.createElement('canvas');c.width=640;c.height=320;const x=c.getContext('2d');x.fillStyle='#f2d18c';x.fillRect(0,0,640,320);x.fillStyle='#bd4d53';x.fillRect(220,100,200,120);return c.toDataURL('image/webp',.8).split(',')[1];});
 const pictures=join(appDataDir,'households/Family/pictures');mkdirSync(pictures,{recursive:true});writeFileSync(join(pictures,id+'-sample.webp'),Buffer.from(webp,'base64'));
 raw=raw.replace('title: Future train',`title: Future train\n      picture: ${id}-sample.webp\n      picture_source: https://images.example/train.png\n      claims: [{user_id: ${userID}}]`);writeFileSync(cfg,raw);await page.goto('/');await expect(card).toHaveClass(/wish-accent-green/);await expect(card.locator('.claimant-name')).toHaveText('Private Claimant');
 const guestContext=await browser.newContext({baseURL,ignoreHTTPSErrors:true}),guest=await guestContext.newPage();await guest.goto(listLinks['Christmas 2001']);
 await expect(guest.locator('.person-summary-total').first()).toHaveText('12.50');await expect(guest.locator('.person-summary-badge.is-claimed').first()).toHaveText(t('P23',1));await expect(guest.locator('.person-summary-badge.is-unclaimed').first()).toHaveText(t('P24',0));expect(await guest.content()).not.toContain('Private Claimant');
 const imageURL=await guest.locator('.wish-picture').first().getAttribute('src');expect((await guest.request.get(imageURL)).headers()['content-type']).toBe('image/webp');
 // Available, missing and absent pictures remain distinct. Public-only errors preserve every draft.
 await card.getByRole('button',{name:t('W13','Future train'),exact:true}).click();modal=page.locator('dialog[open]');await modal.getByLabel(t('I01'),{exact:true}).fill('http://127.0.0.1/private.png');await modal.getByLabel(t('W24'),{exact:true}).fill('  Future train  ');await modal.getByRole('button',{name:t('W14'),exact:true}).click();modal=page.locator('dialog[open]');await expect(modal.getByRole('alert')).toHaveText(t('I09'));await expect(modal.getByLabel(t('W24'),{exact:true})).toHaveValue('  Future train  ');await expect(modal.getByLabel(t('I01'),{exact:true})).toHaveValue('http://127.0.0.1/private.png');await modal.getByLabel(t('I01'),{exact:true}).fill('https://images.example/train.png');await modal.getByRole('button',{name:t('W22'),exact:true}).click();await expect(card).toHaveClass(/wish-accent-green/);
 // @test WSH-006-AC10: retain actual animation and reduced-motion checks.
 const imageModal=page.locator('.picture-modal').first();
 const opening=await card.locator('.wish-picture-button').evaluate(button=>{button.click();return document.querySelector('.picture-modal').getAnimations().map(a=>a.effect.getTiming().easing);});
 expect(opening).toContain('ease-in-out');await expect(imageModal).toBeVisible();
 const closing=await imageModal.evaluate(modal=>{modal.querySelector('[data-close-modal]').click();return {open:modal.open,easing:modal.getAnimations().map(a=>a.effect.getTiming().easing)};});
 expect(closing.open).toBe(true);expect(closing.easing).toContain('ease-in-out');await expect(imageModal).not.toBeVisible();
 await page.emulateMedia({reducedMotion:'reduce'});await card.locator('.wish-picture-button').click();expect(await imageModal.evaluate(m=>m.getAnimations().length)).toBe(0);await imageModal.getByRole('button').click();await expect(imageModal).not.toBeVisible();await page.emulateMedia({reducedMotion:'no-preference'});
 for(const viewport of [{width:1440,height:1000},{width:390,height:844}]){
  await page.setViewportSize(viewport);await expectNightTheme(page);await page.screenshot({path:test.info().outputPath(`household-${viewport.width}.png`),fullPage:true});
  await card.locator('.wish-picture-button').evaluate(button=>button.click());const pictureModal=page.locator('dialog[open]');await expect(pictureModal).toHaveClass(/picture-modal/);await pictureModal.locator('img').click();await expect(pictureModal).toBeVisible();await page.mouse.click(2,2);await expect(pictureModal).toHaveCount(0);
 }
 await assign('Child Name · Birthday 2027');await guest.reload();await expect(guest.locator('.wish-card')).toHaveCount(0);expect((await guest.request.get(imageURL)).status()).toBe(404);
 await guest.goto(listLinks['Birthday 2027']);await expect(guest.locator('.wish-card')).toHaveCount(1);await expect(guest.locator('.person-summary-total')).toHaveText('12.50');
 await assign(t('HR76'));await guest.reload();await expect(guest.locator('.wish-card')).toHaveCount(0);await expect(card.locator('.claimant-name')).toHaveText('Private Claimant');
 // Filtering and remembered sorting keep actions attached to stable wish IDs.
 await page.locator('form[action="/wishes/save"]').last().getByLabel(t('W24'),{exact:true}).fill('Another plain wish');await page.locator('form[action="/wishes/save"]').last().getByRole('button',{name:t('W23'),exact:true}).click();await page.locator('[data-wish-sort]').selectOption('newest');await expect(page.locator('.wish-title')).toHaveText(['Another plain wish','Future train']);await page.reload();await expect(page.locator('[data-wish-sort]')).toHaveValue('newest');
 await page.locator('[name=unassigned]').check();await page.getByRole('button',{name:t('HR89'),exact:true}).click();await expect(page.locator('.wish-card')).toHaveCount(2);await page.getByRole('link',{name:t('HR90'),exact:true}).click();
 await assign('Child Name · Birthday 2027');
 await page.getByRole('link',{name:'Birthday 2027',exact:true}).click();await expect(page.locator('.person-summary-total')).toHaveText('12.50');await page.locator('.person-summary-card').click();const recipientURL=page.url();
 // Creation in this recipient context assigns automatically; duplicate titles are allowed.
 await page.locator('form[action="/wishes/save"]').last().getByLabel(t('W24'),{exact:true}).fill('Recipient gift title');await page.locator('form[action="/wishes/save"]').last().getByRole('button',{name:t('W23'),exact:true}).click();await expect(page).toHaveURL(recipientURL);await expect(page.locator('.wish-card')).toHaveCount(2);await guest.reload();await expect(guest.locator('.wish-card')).toHaveCount(2);
 await page.setViewportSize({width:390,height:844});await expectNightTheme(page);await page.screenshot({path:test.info().outputPath('recipient-mobile.png'),fullPage:true});
 await appControl.restart();await page.reload();await expect(page.locator('.wish-card')).toHaveCount(2);
 // Cancel deletion first; confirming removes only this stable wish and its picture/claims.
 card=page.locator('.wish-card').filter({has:page.getByRole('heading',{name:'Future train',exact:true})});await card.getByRole('button',{name:t('W16','Future train'),exact:true}).click();await page.locator('dialog[open]').getByRole('button',{name:t('G03'),exact:true}).click();expect(existsSync(join(pictures,id+'-sample.webp'))).toBe(true);await card.getByRole('button',{name:t('W16','Future train'),exact:true}).click();await page.locator('dialog[open]').getByRole('button',{name:t('W19'),exact:true}).click();await expect(page.locator('.wish-card')).toHaveCount(1);expect(existsSync(join(pictures,id+'-sample.webp'))).toBe(false);await guest.reload();await expect(guest.locator('.wish-card')).toHaveCount(1);await guestContext.close();
});
