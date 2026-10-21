const {test,expect}=require('./fixtures');
const {expectNightTheme}=require('./night-assertions');
const {readFileSync,writeFileSync,mkdirSync}=require('node:fs');
const {join,dirname}=require('node:path');

// @test WLS-006-AC1 WLS-006-AC2 WLS-006-AC3 WLS-006-AC4 WLS-006-AC5 WLS-006-AC6
for(const locale of ['en-US','de-CH']) test(`Person summaries use exact totals and complete-wish badges in ${locale}`,async({page,context,appDataDir})=>{
 test.setTimeout(60000);
 const de=locale==='de-CH',text=(en,german)=>de?german:en;
 await page.goto('/');await page.getByLabel('Household name').fill('Family');
 await page.getByLabel('Household timezone').fill('Europe/Zurich');
 await page.getByLabel('Language',{exact:true}).selectOption(locale);
 await page.getByRole('button',{name:'Create access links'}).click();
 const family=await page.getByLabel('Family link',{exact:true}).inputValue();
 await page.getByRole('button',{name:'Finish setup'}).click();const parent=page.url();
 const year=Number((await page.getByText(/Parent access|Elternzugang/).textContent()).match(/\d+/)[0]);
 for(const name of ['Anna','MaximilianAlexander','Emma']) {
  await page.getByLabel(text('Person name','Name der Person'),{exact:true}).fill(name);
  await page.getByRole('button',{name:text('Add person','Person hinzufügen'),exact:true}).click();
 }
 await expect(page.getByRole('link',{name:'Emma',exact:true})).toBeVisible();
 const path=name=>join(appDataDir,'households','Family',String(year),'christmas',name,'wishlist.yaml');
 const sample='wishes:\n - title: Complete claimed train\n   price: "10.10"\n   links: [https://shop.example/train]\n   picture: unavailable.webp\n   picture_source: https://never-fetch.example/picture.png\n   claims: [{name: Auntie}, {name: Uncle}]\n - title: Complete unclaimed train\n   price: "2.25"\n   links: [https://shop.example/other]\n - title: Complete zero train\n   price: "0"\n   links: [https://shop.example/free]\n - title: Incomplete priced train\n   price: "5"\n   claims: [{name: Hiddenclaimant}]\n - title: Incomplete title only\n';
 writeFileSync(path('Anna'),sample);
 writeFileSync(path('MaximilianAlexander'),'wishes:\n - title: Another priced wish\n   price: "1024.50"\n   links: [https://shop.example/wish]\n');
 const remoteRequests=[];page.on('request',r=>{if(r.url().includes('never-fetch.example'))remoteRequests.push(r.url());});
 await page.reload();
 const row=(viewer,name)=>viewer.getByRole('link',{name,exact:true});
 async function summary(viewer,name,total,claimed,unclaimed) {
  const person=row(viewer,name);
  await expect(person.locator('.person-summary-total')).toHaveText(total);
  await expect(person.locator('.person-summary-badge.is-claimed')).toHaveText(`${claimed} ${text('claimed','reserviert')}`);
  await expect(person.locator('.person-summary-badge.is-unclaimed')).toHaveText(`${unclaimed} ${text('unclaimed','nicht reserviert')}`);
 }
 await summary(page,'Anna','17.35',1,2);await summary(page,'MaximilianAlexander','1024.50',0,1);await summary(page,'Emma','0.00',0,0);
 await expect(page.locator('.person-summary-help')).toHaveText(text('The counts include only complete wishes. The price total also includes incomplete wishes.','Gezählt werden nur vollständige Wünsche. Die Preissumme enthält auch unvollständige Wünsche.'));
 await expect(row(page,'Anna')).toHaveAccessibleDescription(/17\.35/);
 for(const [screen,width,height] of [['desktop',1440,1000],['mobile',390,844],['narrow',320,740]]) {
  await page.setViewportSize({width,height});await expectNightTheme(page);
  if(width<640) {
   const name=await row(page,'MaximilianAlexander').locator('.person-summary-name').boundingBox();
   const total=await row(page,'MaximilianAlexander').locator('.person-summary-amount').boundingBox();
   expect(total.y).toBeGreaterThanOrEqual(name.y+name.height);
  }
  if(screen!=='narrow') await page.screenshot({path:test.info().outputPath(`overview-${locale}-${screen}.png`),fullPage:true});
 }
 expect(readFileSync(path('Anna'),'utf8')).toBe(sample);
 expect(remoteRequests).toEqual([]);
 const guest=await context.newPage();await guest.goto(family);
 await summary(guest,'Anna','17.35',1,2);
 for(const name of ['Auntie','Uncle','Hiddenclaimant']) expect(await guest.content()).not.toContain(name);
 // Multiple names still count as one claimed wish; claiming a different wish changes the badges on reload.
 await context.addCookies([{name:'wishmas_claim_name',value:'Helper',url:parent}]);
 await row(guest,'Anna').click();
 await guest.getByRole('button',{name:text('Claim wish: Complete unclaimed train','Wunsch reservieren: Complete unclaimed train'),exact:true}).click();
 await page.reload();await summary(page,'Anna','17.35',2,1);
 await guest.getByRole('button',{name:text('Remove my claim: Complete unclaimed train','Meine Reservierung aufheben: Complete unclaimed train'),exact:true}).click();
 await page.reload();await summary(page,'Anna','17.35',1,2);
 // Manual price/completeness edits and large exact sums are reflected on the next overview load.
 writeFileSync(path('Anna'),'wishes:\n - title: Huge priced wish\n   price: "9007199254740992.01"\n - title: Tiny priced wish\n   price: "0.02"\n');
 await page.reload();await summary(page,'Anna','9007199254740992.03',0,0);
 const long='9'.repeat(80)+'.99';
 writeFileSync(path('Anna'),`wishes:\n - title: Very large priced wish\n   price: "${long}"\n - title: Carry priced wish\n   price: "0.01"\n`);
 await page.reload();await summary(page,'Anna','1'+'0'.repeat(80)+'.00',0,0);await expectNightTheme(page);
 const old=join(appDataDir,'households','Family',String(year-1),'christmas','Anna','wishlist.yaml');mkdirSync(dirname(old),{recursive:true});
 const historical='wishes:\n - title: Historical train\n   price: "99.99"\n   links: [https://shop.example/history]\n';writeFileSync(old,historical);
 const history=new URL(parent);history.searchParams.set('year',String(year-1));await page.goto(history.toString());
 await summary(page,'Anna','99.99',0,1);await expect(page.locator('.person-summary-card')).toHaveCount(1);
 expect(readFileSync(old,'utf8')).toBe(historical);
 const denied=new URL(family);denied.searchParams.set('year',String(year-1));expect((await guest.goto(denied.toString())).status()).toBe(404);
 await guest.close();
});
