// @test CON-001-AC1 CON-001-AC2 CON-001-AC3 CON-001-AC4 CON-001-AC5 CON-001-AC6
// @test ACC-006-AC4 ACC-006-AC8 ACC-006-AC10 CON-007-AC4 CON-007-AC5 CON-007-AC6
// @test WSH-006 WLS-010 WLS-011 CLM-005 CLM-007
// Isolated real Traefik HTTPS; never mounts repository household data.
const {execFileSync}=require('node:child_process');
const {mkdtempSync,writeFileSync,readFileSync,rmSync}=require('node:fs');
const {tmpdir}=require('node:os');const {join}=require('node:path');
const assert=require('node:assert/strict');const {chromium,expect}=require('@playwright/test');
const messages=require('../../internal/app/messages.json');
const id=`wishmas-smoke-${process.pid}-${Date.now()}`,name=`${id}-app`,proxy=`${id}-proxy`,network=`${id}-net`,volume=`${id}-data`;
const image=process.env.WISHMAS_SMOKE_IMAGE||`${id}:test`;
const traefik=process.env.WISHMAS_TRAEFIK_IMAGE||'traefik@sha256:31267173a15b4944e797a76ffd9c419707c8d8b32fe5b610f80cd0cfa05f372d';
const docker=(...args)=>execFileSync('docker',args,{encoding:'utf8',stdio:['ignore','pipe','pipe']}).trim();
const endpoint=process.env.DOCKER_HOST||docker('context','inspect','--format','{{.Endpoints.docker.Host}}');
const hostname=/^(tcp|https?):/.test(endpoint)?new URL(endpoint).hostname:'127.0.0.1';
const bind=['127.0.0.1','localhost','[::1]'].includes(hostname)?'127.0.0.1':'0.0.0.0';
const locale=process.env.WISHMAS_SMOKE_LANGUAGE||'en-US';
const t=(key,...args)=>{let i=0;return messages[key][locale==='de-CH'?'de':'en'].replace(/\{[^}]+\}/g,()=>String(args[i++]));};
const fixturePassword='PublicContainerFixture1!';
const uid=process.env.WISHMAS_SMOKE_UID||'1000',gid=process.env.WISHMAS_SMOKE_GID||'1000';
assert.match(uid,/^[0-9]+$/);assert.match(gid,/^[0-9]+$/);
const overrideUser=Boolean(process.env.WISHMAS_SMOKE_UID||process.env.WISHMAS_SMOKE_GID);
const toolsImage='golang:1.26.9-alpine';
const temp=mkdtempSync(join(tmpdir(),'wishmas-container-'));
let browser,origin,trustedIP;
function startApp(){
 docker('run','-d','--name',name,...(overrideUser?['--user',`${uid}:${gid}`]:[]),'--network',network,'--env',`WISHMAS_TRUSTED_PROXIES=${trustedIP}/32`,'--mount',`type=volume,src=${volume},dst=/data`,image);
 const info=JSON.parse(docker('inspect',name))[0];
 assert.equal(info.Config.User,`${uid}:${gid}`);
 assert.ok(Object.values(info.NetworkSettings.Ports||{}).every(value=>value===null),'App backend must have no published port');
}
async function ready(page){for(let i=0;i<100;i++){try{const r=await page.goto(origin,{timeout:2000});if(r.status()===200)return;}catch{}await new Promise(r=>setTimeout(r,100));}throw new Error('HTTPS app did not become ready');}
function saved(){const path=join(temp,'saved.yaml');docker('cp',`${name}:/data/households/SmokeFamily/household.yaml`,path);return readFileSync(path,'utf8');}
(async()=>{try{
 if(!process.env.WISHMAS_SMOKE_IMAGE)execFileSync('docker',['build','-t',image,'.'],{stdio:'inherit'});
 assert.equal(JSON.parse(docker('image','inspect',image))[0].Config.User,'1000:1000','Image default must be 1000:1000');
 // The container includes the offline helper; no private input is supplied here.
 try{docker('run','--rm','--entrypoint','/wishmas-admin',image);assert.fail('Missing command should fail');}catch(error){assert.match(String(error.stderr),/usage: wishmas-admin hash-password \| migrate/);}
 docker('network','create',network);docker('volume','create',volume);
 // Operators provision storage ownership separately from the runtime override.
 // This helper touches a new disposable test volume only; keep it nonempty so
 // Docker does not initialize it again from the image's default /data owner.
 if(overrideUser)docker('run','--rm','--user','0:0','--entrypoint','/bin/sh','--mount',`type=volume,src=${volume},dst=/data`,toolsImage,'-c','chown "$1:$2" /data && touch /data/.ownership-ready','ownership-fixture',uid,gid);
 writeFileSync(join(temp,'dynamic.yml'),`http:\n  routers:\n    wishmas:\n      rule: PathPrefix(\u0060/\u0060)\n      entryPoints: [secure]\n      service: wishmas\n      tls: {}\n  services:\n    wishmas:\n      loadBalancer:\n        servers:\n          - url: http://${name}:8080\n`);
 docker('create','--name',proxy,'--network',network,'-p',`${bind}::443`,'-p',`${bind}::80`,traefik,'--entrypoints.web.address=:80','--entrypoints.web.http.redirections.entrypoint.to=secure','--entrypoints.web.http.redirections.entrypoint.scheme=https','--entrypoints.secure.address=:443','--providers.file.filename=/dynamic.yml','--global.checknewversion=false','--global.sendanonymoususage=false');
 docker('cp',join(temp,'dynamic.yml'),`${proxy}:/dynamic.yml`);docker('start',proxy);
 const proxyInfo=JSON.parse(docker('inspect',proxy))[0];trustedIP=proxyInfo.NetworkSettings.Networks[network].IPAddress;
 origin=`https://${hostname}:${proxyInfo.NetworkSettings.Ports['443/tcp'][0].HostPort}`;
 startApp();browser=await chromium.launch();const admin=await browser.newContext({ignoreHTTPSErrors:true}),user=await browser.newContext({ignoreHTTPSErrors:true});
 const page=await admin.newPage(),guest=await user.newPage();await ready(page);
 const redirect=await admin.request.get(`http://${hostname}:${proxyInfo.NetworkSettings.Ports['80/tcp'][0].HostPort}`,{maxRedirects:0});assert.ok([301,308].includes(redirect.status()));assert.match(redirect.headers().location,/^https:/);
 await page.locator('[name=household]').fill('SmokeFamily');await page.locator('[name=password]').fill(fixturePassword);await page.locator('[name=password_confirm]').fill(fixturePassword);await page.locator('[name=language]').selectOption(locale);await page.getByRole('button',{name:'Finish setup',exact:true}).click();
 const session=(await admin.cookies()).find(cookie=>cookie.name==='__Host-wishmas-admin');assert.ok(session?.secure&&session.httpOnly&&session.sameSite==='Lax');assert.match(session.name,/^__Host-/);
 for(const [kind,value]of [['recipient','Anna'],['user','Aunt Alice'],['user','Uncle Bob'],['list','Christmas 2026']]){await page.locator(`#${kind}-name`).fill(value);await page.locator(`form[action="/manage/${kind}/create"] button`).click();}
 await page.getByRole('link',{name:'Christmas 2026',exact:true}).click();await page.getByRole('button',{name:t('HR50'),exact:true}).click();const links=await page.locator('[data-access-link]').evaluateAll(inputs=>inputs.map(i=>i.value));assert.equal(links.length,2);assert.notEqual(links[0],links[1]);
 await page.locator('.person-summary-card').click();const recipientURL=page.url(),form=page.locator('form[action="/wishes/save"]').last();await form.getByLabel(t('W24'),{exact:true}).fill('Wooden train');await form.getByLabel(t('W10'),{exact:true}).fill('49.90');await form.getByLabel(t('W26'),{exact:true}).fill('https://shop.example/train');
 // Public CA trust and pure-Go scratch encoding; source can be overridden for reruns.
 await form.getByLabel(t('I01'),{exact:true}).fill(process.env.WISHMAS_SMOKE_PICTURE||'https://asset.conrad.com/media10/isa/160267/c1/-/de/003375320II00/image.jpg?x=1000&y=1000&format=jpg&ex=1000&ey=1000&align=center');await form.getByRole('button',{name:t('W23'),exact:true}).click();
 await expect(page.locator('.wish-picture')).toBeVisible({timeout:45000});assert.equal(await page.locator('.wish-picture').evaluate(i=>i.decode().then(()=>Math.max(i.naturalWidth,i.naturalHeight))),640);
 const adminPicture=await page.locator('.wish-picture').getAttribute('src');const denied=await user.request.get(origin+adminPicture);assert.equal(denied.status(),200);assert.match(denied.headers()['content-type'],/^text\/html/);assert.ok((await denied.text()).includes(t('HR07')));
 await guest.goto(links[0]);await guest.getByRole('button',{name:t('C23','Wooden train'),exact:true}).click();await expect(guest.getByText(t('C24'),{exact:true})).toBeVisible();
 await page.reload();await expect(page.locator('.claimant-name')).toHaveText('Aunt Alice');assert.equal(await guest.locator('.claimant-name').count(),0);assert.ok(!(await guest.content()).includes('Uncle Bob'));
 assert.equal(docker('run','--rm','--entrypoint','/bin/sh','--mount',`type=volume,src=${volume},dst=/data,readonly`,toolsImage,'-c','stat -c "%u:%g" /data/households/SmokeFamily/household.yaml'),`${uid}:${gid}`);
 const before=saved();assert.ok(!before.includes(fixturePassword));assert.ok(!before.includes(session.value));assert.match(before,/schema_version: 2/);assert.match(before,/token_hash:/);
 const picture=await admin.request.get(origin+adminPicture);assert.equal(picture.headers()['content-type'],'image/webp');const bytes=await picture.body();
 const persistentCheck=async()=>{await ready(page);await page.goto(recipientURL);await expect(page.locator('.claimant-name')).toHaveText('Aunt Alice');await guest.goto(links[0]);await expect(guest.getByText(t('C24'),{exact:true})).toBeVisible();assert.deepEqual(await (await admin.request.get(origin+adminPicture)).body(),bytes);assert.equal(saved(),before);};
 docker('restart',name);await persistentCheck();docker('rm','-f',name);startApp();await persistentCheck();
 for(const width of [1440,390]){await page.setViewportSize({width,height:900});await page.screenshot({path:join(temp,`admin-${width}.png`),fullPage:true});assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));}
 // A different user shares list visibility but owns no existing claim.
 await guest.goto(links[1]);await expect(guest.getByRole('button',{name:t('C23','Wooden train'),exact:true})).toBeVisible();assert.equal(await guest.locator('.claimant-name').count(),0);
 // A valid user link overrides the existing admin session.
 await page.goto(links[0]);assert.equal(await page.locator('[data-open-modal^="edit-"]').count(),0);await page.goto(recipientURL);
 const afterRestart=await admin.cookies();assert.equal(afterRestart.find(c=>c.name===session.name)?.value,session.value);
 const listURL=new URL(recipientURL);listURL.searchParams.delete('recipient');await page.goto(listURL.href);
 await page.getByRole('button',{name:t('HR72','Aunt Alice','Christmas 2026'),exact:true}).click();await page.locator('dialog[open]').getByRole('button',{name:t('HR75'),exact:true}).click();
 const regenerated=await page.locator('[data-access-link]').first().inputValue();assert.notEqual(regenerated,links[0]);assert.equal(await page.locator('[data-access-link]').nth(1).inputValue(),links[1]);
 await guest.goto(links[0]);await expect(guest.getByRole('heading',{name:t('HR07'),exact:true})).toBeVisible();await guest.goto(regenerated);await expect(guest.getByRole('button',{name:t('C25','Wooden train'),exact:true})).toBeVisible();
 await page.getByRole('button',{name:t('HR14'),exact:true}).click();await expect(page.getByRole('heading',{name:t('HR07'),exact:true})).toBeVisible();
 // Traefik strips untrusted forwarding values; changing a spoofed IP cannot
 // evade the shared client-IP limit. No production insecure-forwarding switch.
 for(let i=0;i<5;i++){
  await admin.setExtraHTTPHeaders({'X-Forwarded-For':`203.0.113.${10+i}`});
  await page.locator('[name=password]').fill('WrongPublicFixture1!');
  await page.getByRole('button',{name:t('HR09'),exact:true}).click();
  await expect(page.getByRole('alert')).toHaveText(t('HR10'));
 }
 await admin.setExtraHTTPHeaders({'X-Forwarded-For':'203.0.113.200'});
 await page.locator('[name=password]').fill(fixturePassword);
 const blocked=page.waitForResponse(r=>r.url().endsWith('/login')&&r.request().method()==='POST');
 await page.getByRole('button',{name:t('HR09'),exact:true}).click();assert.equal((await blocked).status(),429);await expect(page.getByRole('alert')).toHaveText(t('HR11'));
 docker('restart',name);await ready(page);await admin.setExtraHTTPHeaders({});await page.locator('[name=password]').fill(fixturePassword);await page.getByRole('button',{name:t('HR09'),exact:true}).click();await expect(page.getByRole('link',{name:'Christmas 2026',exact:true})).toBeVisible();
 console.log(`PASS ${locale} UID/GID ${uid}:${gid}: real Traefik TLS/HTTP redirect, private nonroot backend, secure hashed sessions, helper, two users, claims/privacy, optimized pictures, byte-identical volume across restart/replacement, logout, spoof-resistant proxy limits/reset and mobile layout.`);
}finally{
 if(browser)await browser.close();
 for(const args of [['rm','-f',name],['rm','-f',proxy],['volume','rm',volume],['network','rm',network],...(!process.env.WISHMAS_SMOKE_IMAGE?[['image','rm',image]]:[])])try{docker(...args);}catch{}
 rmSync(temp,{recursive:true,force:true});
}})().catch(error=>{console.error(error.message);process.exitCode=1;});
