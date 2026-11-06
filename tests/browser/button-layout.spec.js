const {test,expect}=require('./fixtures');
const {readFileSync,writeFileSync}=require('node:fs');
const {join}=require('node:path');
test.use({ignoreHTTPSErrors:true});
const widths=[320,390,640,768,1440];
const password='PublicLayoutFixture1!';

async function seed(page,appDataDir,locale){
 await page.goto('/');await page.locator('[name=household]').fill('LayoutFamily');await page.locator('[name=language]').selectOption(locale);
 await page.locator('[name=password]').fill(password);await page.locator('[name=password_confirm]').fill(password);await page.locator('form[action="/setup/finish"] button').click();
 await expect(page.locator('#list-name')).toBeVisible();
 const file=join(appDataDir,'households','LayoutFamily','household.yaml');
 const a='a'.repeat(64),b='b'.repeat(64);
 const raw=readFileSync(file,'utf8').replace('users: []','users:\n  - {id: u1, name: Alex Example}\n  - {id: u2, name: Morgan Example}').replace('recipients: []','recipients:\n  - {id: r1, name: Anna Example}\n  - {id: r2, name: Ben Example}').replace('lists: []',`lists:\n  - id: l1\n    name: Christmas celebration 2026\n    recipient_ids: [r1]\n    user_links: [{user_id: u1, code: "${a}"}, {user_id: u2, code: "${b}"}]`).replace('wishes: []',`wishes:\n  - id: w1\n    title: Wooden train and winter village\n    price: "49.90"\n    note: A fictional wish with enough detail to wrap on a narrow screen.\n    claims: [{user_id: u1}]\n    assignment: {list_id: l1, recipient_id: r1}\n  - id: w2\n    title: Illustrated storybook collection\n    claims: []\n    assignment: {list_id: l1, recipient_id: r1}\n  - id: w3\n    title: Unassigned winter puzzle\n    claims: []`);
 writeFileSync(file,raw);await page.goto('/');await expect(page.locator('.wish-card')).toHaveCount(3);
 return '/?code='+a;
}

async function expectCreationRows(page,width){
 for(const form of await page.locator('.management-create').all()){
  const input=await form.locator('.management-input').boundingBox(),button=await form.locator('button').boundingBox();
  expect(input).not.toBeNull();expect(button).not.toBeNull();
  if(width<640){
   expect(Math.abs(input.x-button.x)).toBeLessThanOrEqual(1);expect(Math.abs(input.width-button.width)).toBeLessThanOrEqual(1);
   expect(input.width).toBeGreaterThan(200);expect(button.y-input.y-input.height).toBeGreaterThanOrEqual(8);
  }else{
   expect(Math.abs(input.y+input.height/2-button.y-button.height/2)).toBeLessThanOrEqual(1);
   expect(button.x).toBeGreaterThan(input.x+input.width);
  }
 }
 expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
}

async function expectGap(form,selector,minimum){
 const field=await form.locator(selector).last().boundingBox(),button=await form.locator('button.management-primary,button.primary-action').boundingBox();
 expect(field).not.toBeNull();expect(button).not.toBeNull();
 expect(button.y-field.y-field.height).toBeGreaterThanOrEqual(minimum-1);
}

for(const locale of ['en-US','de-CH']){
 // @test CON-003-AC14 CON-003-AC16
 test(`Creation rows align with controls in ${locale}`,async({page,appDataDir})=>{
  await seed(page,appDataDir,locale);
  for(const width of widths){
   await page.setViewportSize({width,height:900});await page.goto('/');await expectCreationRows(page,width);
   await page.goto('/list?id=l1');await expectCreationRows(page,width);
   await page.screenshot({path:test.info().outputPath(`list-${width}.png`),fullPage:true});
  }
 });
 // @test CON-003-AC15 CON-003-AC16
 test(`Form actions keep spacing and disabled semantics in ${locale}`,async({page,appDataDir,browser,baseURL})=>{
  await page.setExtraHTTPHeaders({'Accept-Language':locale});await page.goto('/');
  for(const width of widths){await page.setViewportSize({width,height:900});await expectGap(page.locator('form[action="/setup/finish"]'),'input[type=range]',24);}
  await seed(page,appDataDir,locale);
  for(const width of widths){
   await page.setViewportSize({width,height:900});await expectGap(page.locator('form[action="/wishes/save"]').last(),'textarea[name=note]',12);
   for(const [id,control]of [['edit-w1','textarea[name=note]'],['claim-w1','select'],['assign-w1','select'],['rename-l1','input:not([type=hidden])']]){
    const modal=page.locator('#'+id);await modal.evaluate(d=>d.showModal());await expectGap(modal.locator('form'),control,12);
    await modal.locator('fieldset').evaluate(f=>f.disabled=true);await expect(modal.locator('form button')).toBeDisabled();await expect(modal.locator('[data-close-modal]')).toBeEnabled();
    await modal.locator('[data-close-modal]').click();await expect(modal).not.toBeVisible();await modal.locator('fieldset').evaluate(f=>f.disabled=false);
   }
  }
  await page.locator('[data-open-modal=edit-w1]').click();await page.locator('#edit-w1 [name=title]').fill('Bad');await page.locator('#edit-w1 form button').click();
  await expect(page.locator('dialog[open]')).toHaveCount(1);await expectGap(page.locator('dialog[open] form'),'textarea[name=note]',12);
  const context=await browser.newContext({baseURL,ignoreHTTPSErrors:true,locale});const guest=await context.newPage();
  try{await guest.goto('/');for(const width of widths){await guest.setViewportSize({width,height:900});await expectGap(guest.locator('form[action="/login"]'),'input[name=password]',24);}}finally{await context.close();}
 });
 // @test CON-003-AC16
 test(`Wish footer actions share a baseline in ${locale}`,async({page,appDataDir})=>{
  await seed(page,appDataDir,locale);
  for(const width of widths){
   await page.setViewportSize({width,height:900});
   const rows=await page.locator('.wish-card').evaluateAll(cards=>{
    const groups=[];
    for(const card of cards){const top=card.getBoundingClientRect().top;let group=groups.find(g=>Math.abs(g.top-top)<2);if(!group){group={top,buttons:[]};groups.push(group);}
     for(const b of card.querySelectorAll('.wish-card-footer button')){const r=b.getBoundingClientRect();group.buttons.push({bottom:r.bottom,height:r.height});}}
    return groups;
   });
   for(const row of rows){const bottoms=row.buttons.map(b=>b.bottom);expect(Math.max(...bottoms)-Math.min(...bottoms)).toBeLessThanOrEqual(1);for(const b of row.buttons)expect(b.height).toBeGreaterThanOrEqual(44);}
   expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
   await page.screenshot({path:test.info().outputPath(`footers-${width}.png`),fullPage:true});
  }
 });
}
