const {test,expect}=require('./fixtures');
const {expectNightTheme}=require('./night-assertions');
const {readFileSync,writeFileSync}=require('node:fs');
const {join}=require('node:path');
const messages=require('../../internal/app/messages.json');
test.use({ignoreHTTPSErrors:true});
const password='PublicRosterFixture1!';

// @test ACC-007-AC1–AC10 WLS-007 WLS-008 WLS-009 CON-003-AC5 CON-003-AC10 INT-007
for(const locale of ['en-US','de-CH'])test(`Household lists, rosters and independent links in ${locale}`,async({page,browser,baseURL,context,appControl})=>{
 test.setTimeout(90000);
 const t=(key,...args)=>{let text=messages[key][locale==='de-CH'?'de':'en'];let i=0;return text.replace(/\{[^}]+\}/g,()=>String(args[i++]));};
 await page.goto('/');await page.locator('[name=household]').fill('Family');await page.locator('[name=language]').selectOption(locale);
 await page.locator('[name=password]').fill(password);await page.locator('[name=password_confirm]').fill(password);await page.getByRole('button',{name:'Finish setup',exact:true}).click();
 await expect(page.getByText(t('HR15'),{exact:true})).toBeVisible();
 const create=async(kind,name)=>{await page.locator(`#${kind}-name`).fill(name);await page.locator(`form[action="/manage/${kind}/create"] button`).click();await expect(page.locator(`#${kind}-name`)).toHaveValue('');};
 await create('user','Grand Mum');await create('recipient','Grand Mum');await create('recipient','Other Recipient');await create('list','Christmas 2001');await create('list','Birthday 2027');await create('user','Second User');
 // Invalid drafts remain, no duplicate entity is created.
 await page.locator('#recipient-name').fill(' grand mum ');await page.locator('form[action="/manage/recipient/create"] button').click();
 await expect(page.getByRole('alert')).toHaveText(t('HR56'));await expect(page.locator('#recipient-name')).toHaveValue('grand mum');
 // Modal editing preserves a rejected draft, with its error inside the dialog.
 await page.getByRole('button',{name:t('HR60')+' · Grand Mum',exact:true}).click();
 let dialog=page.getByRole('dialog');await dialog.locator('input[name=name]').fill('ab');await dialog.getByRole('button',{name:t('HR61'),exact:true}).click();
 dialog=page.getByRole('dialog');await expect(dialog).toBeVisible();await expect(dialog.getByRole('alert')).toHaveText(t('HR66'));await expect(dialog.locator('input[name=name]')).toHaveValue('ab');
 await dialog.locator('input[name=name]').fill('Renamed User');await dialog.getByRole('button',{name:t('HR61'),exact:true}).click();await expect(page.getByRole('dialog')).toHaveCount(0);
 // Cancellation is a pure UI operation.
 await page.getByRole('button',{name:t('HR35')+' · Christmas 2001',exact:true}).click();await expect(page.getByRole('dialog')).toContainText(t('HR38',0,0));await page.getByRole('dialog').getByRole('button',{name:t('G03'),exact:true}).click();await expect(page.getByRole('link',{name:'Christmas 2001',exact:true})).toBeVisible();
 await page.getByRole('link',{name:'Christmas 2001',exact:true}).click();
 const add=async name=>{await page.locator('#recipient-select').selectOption({label:name});await page.getByRole('button',{name:t('HR50'),exact:true}).click();};
 await add('Other Recipient');await add('Grand Mum');
 // @test WLS-011-AC7 WLS-011-AC8: summaries and removal share the first panel.
 const panels=page.locator('main > section.management-panel');
 await expect(panels).toHaveCount(2);await expect(panels.locator(':scope > h2')).toHaveText([t('HR21'),t('HR68')]);
 await expect(panels.first().locator('.person-summary-card')).toHaveCount(2);
 await expect(page.locator('.management-row > a')).toHaveCount(0);
 const recipientCard=panels.first().locator('li').filter({has:page.getByRole('link',{name:'Grand Mum',exact:true})});
 await expect(recipientCard.getByRole('button',{name:t('HR52')+' · Grand Mum',exact:true})).toBeVisible();
 await expect(recipientCard.locator('a button')).toHaveCount(0);
 // Household added order wins over the order in which recipients were selected.
 await expect(page.locator('.person-summary-name')).toHaveText(['Grand Mum','Other Recipient']);
 const inputs=page.locator('[data-access-link]');await expect(inputs).toHaveCount(2);
 const oldLink=await inputs.nth(0).inputValue(),secondLink=await inputs.nth(1).inputValue();expect(oldLink.startsWith(baseURL+'/?code=')).toBe(true);expect(oldLink).not.toBe(secondLink);
 // Clipboard uses the existing icon/checkmark, restoring the icon after three seconds.
 await context.grantPermissions(['clipboard-read','clipboard-write']);const copy=page.getByRole('button',{name:t('HR71','Renamed User','Christmas 2001'),exact:true});
 await copy.click();await expect(copy).toHaveAttribute('data-copied','true');expect(await page.evaluate(()=>navigator.clipboard.readText())).toBe(oldLink);await expect(copy).toHaveAttribute('data-copied','false',{timeout:5000});
 await page.setViewportSize({width:1440,height:1000});await expectNightTheme(page);await page.screenshot({path:test.info().outputPath('list-desktop.png'),fullPage:true});
 const a=await browser.newContext({baseURL,ignoreHTTPSErrors:true}),b=await browser.newContext({baseURL,ignoreHTTPSErrors:true});const first=await a.newPage(),second=await b.newPage();
 await first.goto(oldLink);await second.goto(secondLink);
 for(const p of [first,second]){await expect(p.getByRole('link',{name:'Grand Mum',exact:true})).toBeVisible();await expect(p.getByRole('link',{name:'Other Recipient',exact:true})).toBeVisible();await expect(p.locator('form')).toHaveCount(0);await expect(p.getByText('Birthday 2027',{exact:true})).toHaveCount(0);}
 await expect(first.getByText(t('HR16','Renamed User'),{exact:true})).toBeVisible();await expect(second.getByText(t('HR16','Second User'),{exact:true})).toBeVisible();
 // Even the signed-in parent sees a restricted view through a user link.
 const listURL=page.url();await page.goto(oldLink);await expect(page.locator('form')).toHaveCount(0);await page.goto(listURL);
 await page.getByRole('button',{name:t('HR72','Renamed User','Christmas 2001'),exact:true}).click();await expect(page.getByRole('dialog')).toContainText(t('HR74','Renamed User'));await page.getByRole('dialog').getByRole('button',{name:t('HR75'),exact:true}).click();
 const newLink=await page.locator('[data-access-link]').first().inputValue();expect(newLink).not.toBe(oldLink);expect(await page.locator('[data-access-link]').nth(1).inputValue()).toBe(secondLink);
 await first.reload();await expect(first.getByRole('heading',{name:t('HR07'),exact:true})).toBeVisible();await expect(first.locator('html')).toHaveAttribute('lang',locale);await second.reload();await expect(second.getByRole('heading',{name:'Christmas 2001',exact:true})).toBeVisible();await first.goto(newLink);
 await first.getByRole('link',{name:'Grand Mum',exact:true}).click();await expect(first.getByRole('link',{name:'Other Recipient',exact:true})).toHaveCount(0);await first.getByRole('link',{name:t('HR25'),exact:true}).click();
 await page.setViewportSize({width:390,height:844});await expectNightTheme(page);await page.screenshot({path:test.info().outputPath('list-mobile.png'),fullPage:true});
 const beforeRemove=page.url();await recipientCard.getByRole('button',{name:t('HR52')+' · Grand Mum',exact:true}).click();await page.getByRole('dialog').getByRole('button',{name:t('G03'),exact:true}).click();await expect(page).toHaveURL(beforeRemove);await expect(recipientCard).toBeVisible();
 await recipientCard.getByRole('button',{name:t('HR52')+' · Grand Mum',exact:true}).click();await expectNightTheme(page);await page.screenshot({path:test.info().outputPath('confirmation-mobile.png'),fullPage:true});await page.getByRole('dialog').getByRole('button',{name:t('HR52'),exact:true}).click();
 await first.reload();await expect(first.getByRole('link',{name:'Grand Mum',exact:true})).toHaveCount(0);await expect(first.getByRole('link',{name:'Other Recipient',exact:true})).toBeVisible();
 await page.getByRole('link',{name:t('HR24'),exact:true}).click();await expectNightTheme(page);await page.screenshot({path:test.info().outputPath('household-mobile.png'),fullPage:true});
 // Rename preserves the list code and changes the next read; old dates remain editable.
 await page.getByRole('button',{name:t('HR33')+' · Christmas 2001',exact:true}).click();await page.getByRole('dialog').locator('input[name=name]').fill('Renamed Christmas');await page.getByRole('dialog').getByRole('button',{name:t('HR34'),exact:true}).click();await first.reload();await expect(first.getByRole('heading',{name:'Renamed Christmas',exact:true})).toBeVisible();
 await appControl.restart();await second.reload();await expect(second.getByRole('heading',{name:'Renamed Christmas',exact:true})).toBeVisible();await page.reload();
 await page.getByRole('link',{name:'Birthday 2027',exact:true}).click();const otherListLink=await page.locator('[data-access-link]').first().inputValue();await page.getByRole('link',{name:t('HR24'),exact:true}).click();
 await page.getByRole('button',{name:t('HR35')+' · Renamed Christmas',exact:true}).click();await page.getByRole('dialog').getByRole('button',{name:t('HR35'),exact:true}).click();await second.reload();await expect(second.getByRole('heading',{name:t('HR07'),exact:true})).toBeVisible();await first.goto(otherListLink);await expect(first.getByRole('heading',{name:'Birthday 2027',exact:true})).toBeVisible();
 await page.getByRole('button',{name:t('HR62')+' · Renamed User',exact:true}).click();await expect(page.getByRole('dialog')).toContainText(t('HR65',0));await page.getByRole('dialog').getByRole('button',{name:t('HR62'),exact:true}).click();await first.reload();await expect(first.getByRole('heading',{name:t('HR07'),exact:true})).toBeVisible();
 await page.getByRole('button',{name:t('HR47')+' · Grand Mum',exact:true}).click();await page.getByRole('dialog').getByRole('button',{name:t('HR47'),exact:true}).click();await expect(page.getByRole('button',{name:t('HR47')+' · Grand Mum',exact:true})).toHaveCount(0);
 await a.close();await b.close();
});

// @test CON-004-AC3 CON-004-AC4 ACC-007-AC4 WLS-009-AC4
 test('Stale confirmations and scoped seeded wishes remain safe',async({page,browser,baseURL,appDataDir})=>{
 await page.goto('/');await page.locator('[name=household]').fill('Family');await page.locator('[name=language]').selectOption('en-US');await page.locator('[name=password]').fill(password);await page.locator('[name=password_confirm]').fill(password);await page.getByRole('button',{name:'Finish setup',exact:true}).click();
 for(const [kind,name] of [['recipient','Child Name'],['user','Other User'],['user','Private Claimant'],['list','Christmas 2001']]){await page.locator(`#${kind}-name`).fill(name);await page.locator(`form[action="/manage/${kind}/create"] button`).click();}
 await page.getByRole('link',{name:'Christmas 2001',exact:true}).click();await page.getByRole('button',{name:'Add recipient to list',exact:true}).click();const link=await page.locator('[data-access-link]').first().inputValue();
 const cfg=join(appDataDir,'households/Family/household.yaml');let raw=readFileSync(cfg,'utf8');const listID=raw.match(/lists:\n\s+- id: (\w+)/)[1],recipientID=raw.match(/recipients:\n\s+- id: (\w+)/)[1],userID=[...raw.split('users:\n')[1].split('recipients:')[0].matchAll(/- id: (\w+)/g)][1][1];
 raw=raw.replace('wishes: []',`wishes:\n    - id: seeded1\n      title: Visible seeded wish\n      price: "12.50"\n      links: ["https://shop.example/product"]\n      assignment: {list_id: ${listID}, recipient_id: ${recipientID}}\n      claims: [{user_id: ${userID}}]\n    - id: seeded2\n      title: Hidden unassigned wish\n      claims: [{user_id: ${userID}}]`);writeFileSync(cfg,raw);
 const userContext=await browser.newContext({baseURL,ignoreHTTPSErrors:true}),guest=await userContext.newPage();await guest.goto(link);await expect(guest.getByText('Visible seeded wish',{exact:true})).toBeVisible();await expect(guest.getByText('Hidden unassigned wish',{exact:true})).toHaveCount(0);expect(await guest.content()).not.toContain('Private Claimant');
 await page.goto('/');await page.getByRole('button',{name:'Delete list · Christmas 2001',exact:true}).click();await expect(page.getByRole('dialog')).toContainText('Wishes to unassign: 1. Claims to remove: 1.');
 // An external edit after confirmation was opened must prevent deleting against stale counts.
 writeFileSync(cfg,readFileSync(cfg,'utf8').replace('Visible seeded wish','Updated seeded wish'));const stable=readFileSync(cfg,'utf8');await page.getByRole('dialog').getByRole('button',{name:'Delete list',exact:true}).click();await expect(page.getByRole('alert')).toContainText('Household data changed.');expect(readFileSync(cfg,'utf8')).toBe(stable);
 await page.getByRole('button',{name:'Delete list · Christmas 2001',exact:true}).click();await page.getByRole('dialog').getByRole('button',{name:'Delete list',exact:true}).click();const saved=readFileSync(cfg,'utf8');expect(saved).toContain('Updated seeded wish');expect(saved).toContain('Hidden unassigned wish');expect(saved).not.toContain('assignment:');expect((saved.match(/user_id:/g)||[]).length).toBe(1);await guest.reload();await expect(guest.getByRole('heading',{name:'Admin sign-in',exact:true})).toBeVisible();await userContext.close();
});

// @test CON-003-AC5 CON-003-AC10: creation actions align with their inputs.
for(const locale of ['en-US','de-CH'])test(`Household creation controls align in ${locale}`,async({page})=>{
 await page.goto('/');await page.locator('[name=household]').fill('AlignmentFamily');await page.locator('[name=language]').selectOption(locale);
 await page.locator('[name=password]').fill(password);await page.locator('[name=password_confirm]').fill(password);await page.getByRole('button',{name:'Finish setup',exact:true}).click();
 for(const width of [320,390,640,1440]){
  await page.setViewportSize({width,height:900});
  for(const form of await page.locator('.management-create').all()){
   const input=await form.locator('.management-input').boundingBox(),button=await form.locator('button').boundingBox();
   if(width<640){
    expect(Math.abs(input.x-button.x)).toBeLessThanOrEqual(1);
    expect(Math.abs(input.width-button.width)).toBeLessThanOrEqual(1);
    expect(button.y).toBeGreaterThanOrEqual(input.y+input.height);
    expect(input.width).toBeGreaterThan(200);
   }else{
    expect(Math.abs(input.y+input.height/2-button.y-button.height/2)).toBeLessThanOrEqual(1);
    expect(button.x).toBeGreaterThanOrEqual(input.x+input.width);
   }
  }
  expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
  await page.screenshot({path:test.info().outputPath(`household-alignment-${locale}-${width}.png`),fullPage:true});
 }
});
