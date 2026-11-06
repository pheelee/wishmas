const {test:base,expect}=require('./fixtures');
const {createServer}=require('node:https');
const {mkdtempSync,readFileSync,rmSync}=require('node:fs');
const {tmpdir}=require('node:os');
const {join}=require('node:path');
const {execFile}=require('node:child_process');
const {promisify}=require('node:util');
const messages=require('../../internal/app/messages.json');
// Disposable protocol fixture; signed-request compatibility is separately tested
// against an actual S3-compatible provider by s3_provider_test.go.
const test=base.extend({
 s3:async({},use)=>{
  const dir=mkdtempSync(join(tmpdir(),'wishmas-s3-tls-'));
  const key=join(dir,'key.pem'),cert=join(dir,'cert.pem');
  try {
  await promisify(execFile)('openssl',['req','-x509','-newkey','rsa:2048','-nodes','-keyout',key,'-out',cert,'-days','1','-subj','/CN=localhost','-addext','subjectAltName=IP:127.0.0.1,DNS:localhost']);
  const state={status:0,objects:new Map(),puts:0,caFile:cert};
  const server=createServer({key:readFileSync(key),cert:readFileSync(cert)},async(req,res)=>{
   if(state.status){res.writeHead(state.status);res.end();return;}
   const u=new URL(req.url,'https://fixture');
   if(u.searchParams.get('list-type')==='2'){res.setHeader('Content-Type','application/xml');res.end('<ListBucketResult>'+[...state.objects.keys()].map(key=>'<Contents><Key>'+key+'</Key></Contents>').join('')+'</ListBucketResult>');return;}
   const key=decodeURIComponent(u.pathname.slice('/bucket/'.length));
   if(req.method==='PUT'){const chunks=[];for await(const c of req)chunks.push(c);state.objects.set(key,Buffer.concat(chunks));state.puts++;res.end();return;}
   if(!state.objects.has(key)){res.writeHead(404);res.end();return;}res.end(state.objects.get(key));
  });
  await new Promise(resolve=>server.listen(0,'127.0.0.1',resolve));state.endpoint=`https://127.0.0.1:${server.address().port}`;
  try{await use(state);}finally{await new Promise(resolve=>server.close(resolve));}
  } finally {rmSync(dir,{recursive:true,force:true});}
 },
 appEnvironment:async({s3},use)=>use({SSL_CERT_FILE:s3.caFile,WISHMAS_STORAGE:'s3',WISHMAS_S3_ENDPOINT:s3.endpoint,WISHMAS_S3_BUCKET:'bucket',WISHMAS_S3_ACCESS_KEY_ID:'fixture',WISHMAS_S3_SECRET_ACCESS_KEY:'PublicFixtureSecret1',WISHMAS_S3_TIMEOUT:'1s'}),
});
test.use({ignoreHTTPSErrors:true});
// @test CON-009-AC6 CON-009-AC9 CON-009-AC12 ACC-008-AC1 ACC-008-AC2 ACC-008-AC3 ACC-008-AC4 ACC-008-AC5 ACC-008-AC8 ACC-008-AC10
for(const locale of ['en-US','de-CH'])test(`S3 cached controls, scoped claims and durable logout in ${locale}`,async({page,browser,baseURL,appControl,s3})=>{
 test.setTimeout(90000);const t=key=>messages[key][locale==='de-CH'?'de':'en'];const password='PublicBrowserFixture1!';
 await page.goto('/');await page.locator('[name=household]').fill('Family');await page.locator('[name=language]').selectOption(locale);await page.locator('[name=password]').fill(password);await page.locator('[name=password_confirm]').fill(password);await page.getByRole('button',{name:'Finish setup',exact:true}).click();
 for(const [kind,name]of [['recipient','Child Name'],['recipient','Other Child'],['user','Aunt Alice'],['user','Uncle Bob'],['list','Christmas 2026']]){await page.locator(`#${kind}-name`).fill(name);await page.locator(`form[action="/manage/${kind}/create"] button`).click();}
 await page.getByRole('link',{name:'Christmas 2026',exact:true}).click();await page.locator('#recipient-select').selectOption({label:'Child Name'});await page.getByRole('button',{name:t('HR50'),exact:true}).click();
 const listURL=page.url(),links=await page.locator('[data-access-link]').evaluateAll(inputs=>inputs.map(i=>i.value));
 await page.getByRole('link',{name:'Child Name',exact:true}).click();const form=page.locator('form[action="/wishes/save"]').last();await form.locator('[name=title]').fill('Wooden train');await form.locator('[name=price]').fill('12.50');await form.locator('[name=links]').fill('https://shop.example/train');await form.getByRole('button',{name:t('W23'),exact:true}).click();
 const guestContext=await browser.newContext({baseURL,ignoreHTTPSErrors:true}),guest=await guestContext.newPage();await guest.goto(links[0]);await guest.locator('.person-summary-card').first().click();await guest.locator('.wish-claim-button').click();await expect(guest.locator('.wish-claim-button')).toHaveAttribute('aria-pressed','true');
 const cookies=await page.context().cookies();const session=cookies.find(c=>c.name==='__Host-wishmas-admin');
 s3.status=503;await page.goto(listURL);await expect(page.locator('[data-storage-readonly]')).toHaveText(t('HR113'));
 await expect(page.locator('form[action="/manage/selection/add"] button')).toBeDisabled();await expect(page.locator('[data-open-modal^=regenerate-]').first()).toBeDisabled();
 const copy=page.locator('[data-copy]').first();await expect(copy).toBeEnabled();await copy.click();
 await page.setViewportSize({width:390,height:844});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);expect(await page.locator('[data-storage-readonly]').evaluate(banner=>banner.getBoundingClientRect().top>=document.querySelector('.light-strand').getBoundingClientRect().bottom)).toBe(true);await page.screenshot({path:test.info().outputPath(`s3-readonly-${locale}-mobile.png`),fullPage:true});
 await guest.reload();await expect(guest.locator('.wish-claim-button')).toBeDisabled();await expect(guest.locator('.wish-card')).toHaveCount(1);expect(await guest.content()).not.toContain('Uncle Bob');await expect(guest.locator('[data-wish-sort]')).toBeEnabled();
 const anonymous=await browser.newContext({baseURL,ignoreHTTPSErrors:true}),login=await anonymous.newPage();await login.goto('/');await expect(login.locator('form[action="/login"] button')).toBeDisabled();
 // Cached pages and local revocation survive a genuine process restart.
 await appControl.restart();await page.goto(listURL);await expect(page.locator('[data-storage-readonly]')).toBeVisible();await expect(page.getByRole('button',{name:t('HR14'),exact:true})).toBeEnabled();await page.getByRole('button',{name:t('HR14'),exact:true}).click();await expect(page.getByRole('heading',{name:t('HR07'),exact:true})).toBeVisible();
 await page.context().addCookies([session]);await appControl.restart();await page.goto('/');await expect(page.getByRole('heading',{name:t('HR07'),exact:true})).toBeVisible();
 s3.status=0;await guest.reload();await expect(guest.locator('[data-storage-readonly]')).toHaveCount(0);await expect(guest.locator('.wish-claim-button')).toBeEnabled();await guest.locator('.wish-claim-button').click();await expect(guest.locator('.wish-claim-button')).toHaveAttribute('aria-pressed','false');
 await page.reload();await page.locator('[name=password]').fill(password);await page.getByRole('button',{name:t('HR09'),exact:true}).click();await expect(page.locator('form[action="/logout"]')).toBeVisible();
 await guestContext.close();await anonymous.close();
});
