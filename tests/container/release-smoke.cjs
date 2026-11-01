// @test REL-001-AC2 REL-002-AC2 REL-002-AC4 CON-001-AC1 CON-001-AC2
// Focused real-image release check; only disposable Docker resources are used.
const {execFileSync}=require('node:child_process');
const {mkdtempSync,writeFileSync,readFileSync,rmSync}=require('node:fs');
const {tmpdir}=require('node:os');
const {join}=require('node:path');
const assert=require('node:assert/strict');
const {chromium,expect}=require('@playwright/test');
const image=process.env.WISHMAS_SMOKE_IMAGE;
assert.ok(image,'Set WISHMAS_SMOKE_IMAGE to the already-built amd64 release image');
const id=`wishmas-release-smoke-${process.pid}-${Date.now()}`;
const name=`${id}-app`,proxy=`${id}-proxy`,network=`${id}-net`,volume=`${id}-data`;
// @impl DEP-001-AC1: stable proxy image tracked by Renovate.
const traefik='traefik:3.7.17@sha256:30ff62491aacdb862ceb8c9edcf2627da638d7deafdd4b6dab32fe34f8aa3e90';
const docker=(...args)=>execFileSync('docker',args,{encoding:'utf8',stdio:['ignore','pipe','pipe']}).trim();
const endpoint=process.env.DOCKER_HOST||docker('context','inspect','--format','{{.Endpoints.docker.Host}}');
const hostname=/^(tcp|https?):/.test(endpoint)?new URL(endpoint).hostname:'127.0.0.1';
const bind=['127.0.0.1','localhost','[::1]'].includes(hostname)?'127.0.0.1':'0.0.0.0';
const temp=mkdtempSync(join(tmpdir(),'wishmas-release-smoke-'));
const password='PublicReleaseFixture1!';
let browser,origin,stage='start';
// @test REL-002-AC2 REL-002-AC4: real HTTPS uses Playwright's Node HTTP
// transport. Chromium only parses inert response HTML, never navigates the network.
const locations=new WeakMap();
async function responsePage(page,response){
 assert.equal(response.status(),200,`HTTP ${response.status()} from ${response.url()}`);
 locations.set(page,response.url());
 await page.setContent(await response.text());
}
async function load(page,url){
 await responsePage(page,await page.request.get(new URL(url,origin).href));
}
async function follow(page,link){
 await load(page,new URL(await link.getAttribute('href'),locations.get(page)).href);
}
async function submit(page,button){
 const submission=await button.evaluate(button=>{
  const form=button.closest('form');
  const values=Object.fromEntries(new FormData(form));
  if(button.name)values[button.name]=button.value;
  return {action:form.getAttribute('action'),values};
 });
 const url=new URL(submission.action,locations.get(page)).href;
 await responsePage(page,await page.request.post(url,{form:submission.values,headers:{Origin:new URL(url).origin}}));
}
async function ready(page){
 const deadline=Date.now()+30000;
 while(Date.now()<deadline){
  try{const response=await page.request.get(origin,{timeout:2000});if(response.status()===200){await responsePage(page,response);return;}}catch{}
  await new Promise(resolve=>setTimeout(resolve,100));
 }
 throw new Error('HTTPS app did not become ready within 30 seconds');
}
function saved(){
 const path=join(temp,'household.yaml');
 docker('cp',`${name}:/data/households/ReleaseFamily/household.yaml`,path);
 return readFileSync(path,'utf8');
}
(async()=>{try{
 stage='container and HTTPS startup';
 assert.equal(JSON.parse(docker('image','inspect',image))[0].Config.User,'1000:1000');
 docker('network','create',network);docker('volume','create',volume);
 writeFileSync(join(temp,'dynamic.yml'),`http:\n  routers:\n    wishmas:\n      rule: PathPrefix(\u0060/\u0060)\n      entryPoints: [secure]\n      service: wishmas\n      tls: {}\n  services:\n    wishmas:\n      loadBalancer:\n        servers:\n          - url: http://${name}:8080\n`);
 docker('create','--name',proxy,'--network',network,'-p',`${bind}::443`,traefik,'--entrypoints.secure.address=:443','--providers.file.filename=/dynamic.yml','--global.checknewversion=false','--global.sendanonymoususage=false');
 docker('cp',join(temp,'dynamic.yml'),`${proxy}:/dynamic.yml`);docker('start',proxy);
 const info=JSON.parse(docker('inspect',proxy))[0];
 const trustedIP=info.NetworkSettings.Networks[network].IPAddress;
 origin=`https://${hostname}:${info.NetworkSettings.Ports['443/tcp'][0].HostPort}`;
 docker('run','-d','--name',name,'--network',network,'--env',`WISHMAS_TRUSTED_PROXIES=${trustedIP}/32`,'--mount',`type=volume,src=${volume},dst=/data`,image);
 browser=await chromium.launch();
 const admin=await browser.newContext({ignoreHTTPSErrors:true,javaScriptEnabled:false});
 const guestContext=await browser.newContext({ignoreHTTPSErrors:true,javaScriptEnabled:false});
 // Block every browser resource request: only the Node HTTPS transport talks to the app.
 await admin.route('**/*',route=>route.abort());
 await guestContext.route('**/*',route=>route.abort());
 const page=await admin.newPage(),guest=await guestContext.newPage();
 await ready(page);
 stage='setup and password login';
 await page.locator('[name=household]').fill('ReleaseFamily');
 await page.locator('[name=password]').fill(password);
 await page.locator('[name=password_confirm]').fill(password);
 await page.locator('[name=language]').selectOption('en-US');
 await submit(page,page.getByRole('button',{name:'Finish setup',exact:true}));
 await submit(page,page.getByRole('button',{name:'Log out',exact:true}));
 await page.locator('[name=password]').fill(password);
 await submit(page,page.getByRole('button',{name:'Sign in',exact:true}));
 const session=(await admin.cookies()).find(cookie=>cookie.name==='__Host-wishmas-admin');
 assert.ok(session?.secure&&session.httpOnly);
 stage='wish creation';
 for(const [kind,value]of [['recipient','Anna'],['user','Aunt Alice'],['list','Christmas 2026']]){
  await page.locator(`#${kind}-name`).fill(value);
  await submit(page,page.locator(`form[action="/manage/${kind}/create"] button`));
 }
 await follow(page,page.getByRole('link',{name:'Christmas 2026',exact:true}));
 const messages=require('../../internal/app/messages.json');
 const t=key=>messages[key].en;
 await submit(page,page.getByRole('button',{name:t('HR50'),exact:true}));
 await expect(page.locator('[data-access-link]')).toHaveCount(1);
 const userLink=await page.locator('[data-access-link]').inputValue();
 await follow(page,page.locator('.person-summary-card'));
 const recipientURL=locations.get(page),form=page.locator('form[action="/wishes/save"]').last();
 await form.getByLabel(t('W24'),{exact:true}).fill('Wooden train');
 await form.getByLabel(t('W10'),{exact:true}).fill('49.90');
 await form.getByLabel(t('W26'),{exact:true}).fill('https://shop.example/train');
 await submit(page,form.getByRole('button',{name:t('W23'),exact:true}));
 stage='claim';
 await load(guest,userLink);
 await submit(guest,guest.locator('.wish-claim-button'));
 await expect(guest.getByText(t('C24'),{exact:true})).toBeVisible();
 await load(page,recipientURL);
 await expect(page.locator('.claimant-name')).toHaveText('Aunt Alice');
 const before=saved();
 stage='restart and persistence';
 docker('restart',name);await ready(page);
 await load(page,recipientURL);
 await expect(page.locator('.claimant-name')).toHaveText('Aunt Alice');
 await load(guest,userLink);
 await expect(guest.getByText(t('C24'),{exact:true})).toBeVisible();
 assert.equal(saved(),before,'Wish, claim and access data must survive restart unchanged');
 console.log('PASS focused release smoke: HTTPS startup, password login, create/claim wish and restart persistence.');
}finally{
 if(browser)await browser.close();
 for(const args of [['rm','-f',name],['rm','-f',proxy],['volume','rm',volume],['network','rm',network]]){
  try{docker(...args);}catch{}
 }
 rmSync(temp,{recursive:true,force:true});
}})().catch(error=>{console.error(`Release smoke failed during ${stage}:`,error.stack||error);process.exitCode=1;});
