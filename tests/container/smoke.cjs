// @test CON-001-AC1 CON-001-AC2 CON-001-AC3 CON-001-AC4 CON-004
// Run with: node tests/container/smoke.cjs (requires Docker and Playwright Chromium).
const { execFileSync } = require('node:child_process');
const { chromium, expect } = require('@playwright/test');
const assert = require('node:assert/strict');
const id = `wishmas-smoke-${process.pid}-${Date.now()}`;
const image = `${id}:test`, volume = `${id}-data`, name = id;
const docker = (...args) => execFileSync('docker', args, { encoding:'utf8', stdio:['ignore','pipe','pipe'] }).trim();
// Remote development daemons publish ports in their own network namespace.
const endpoint = process.env.DOCKER_HOST || docker('context','inspect','--format','{{.Endpoints.docker.Host}}');
const hostname = /^(tcp|https?):/.test(endpoint) ? new URL(endpoint).hostname : '127.0.0.1';
const local = ['127.0.0.1','localhost','[::1]'].includes(hostname);
const german = process.env.WISHMAS_SMOKE_LANGUAGE === 'de-CH';
const text = (en, de) => german ? de : en;
let browser;
async function start() {
 docker('run','-d','--name',name,'-p',`${local ? '127.0.0.1' : '0.0.0.0'}::8080`,'--mount',`type=volume,src=${volume},dst=/data`,image);
 return await ready();
}
async function ready() {
 const binding = JSON.parse(docker('inspect',name))[0].NetworkSettings.Ports['8080/tcp'][0];
 const origin = `http://${hostname}:${binding.HostPort}`;
 for(let i=0;i<100;i++) {
  try { await fetch(origin,{signal:AbortSignal.timeout(1000)}); return origin; }
  catch(error) { if(i===99) { console.error(docker('logs',name)); throw error; } await new Promise(resolve=>setTimeout(resolve,100)); }
 }
}
(async()=>{
 try {
  console.log('Building smoke-test image…');
  execFileSync('docker',['build','-t',image,'.'],{stdio:'inherit'});
  docker('volume','create',volume);
  let origin=await start();
  browser=await chromium.launch();
  const parents=await browser.newContext(), familyContext=await browser.newContext();
  const page=await parents.newPage(), guest=await familyContext.newPage();
  await page.goto(origin);
  await page.getByLabel('Household name').fill('SmokeFamily');
  await page.getByLabel('Household timezone').fill('Europe/Zurich');
  await page.getByLabel('Language',{exact:true}).selectOption(german ? 'de-CH' : 'en-US');
  await page.getByRole('button',{name:'Create access links'}).click();
  const familyPath = new URL(await page.getByLabel('Family link',{exact:true}).inputValue()).search;
  await page.getByRole('button',{name:'Finish setup'}).click();
  const parentPath=new URL(page.url()).search;
  await page.getByLabel(text('Person name','Name der Person'),{exact:true}).fill('Anna');
  await page.getByRole('button',{name:text('Add person','Person hinzufügen'),exact:true}).click();
  await page.getByRole('link',{name:'Anna',exact:true}).click();
  const add=page.getByRole('region',{name:text('Add wish','Wunsch hinzufügen'),exact:true});
  await add.getByLabel(text('Title','Titel'),{exact:true}).fill('Wooden train');
  await add.getByLabel(text('Reference price','Richtpreis'),{exact:true}).fill('49.90');
  await add.getByRole('button',{name:text('Add wish','Wunsch hinzufügen'),exact:true}).click();
  await guest.goto(origin+'/'+familyPath);
  await guest.getByRole('link',{name:'Anna',exact:true}).click();
  await guest.getByRole('button',{name:text('Claim wish: Wooden train','Wunsch reservieren: Wooden train'),exact:true}).click();
  await guest.getByLabel(text('Claim name','Name für Reservierungen'),{exact:true}).fill('Auntie');
  await guest.getByRole('button',{name:text('Save name and claim','Namen speichern und reservieren'),exact:true}).click();
  await expect(guest.getByText(text('Claimed by you','Von dir reserviert'),{exact:true})).toBeVisible();
  await expect(guest.getByRole('button',{name:/^(Edit wish:|Delete wish:|Remove claim by|Wunsch bearbeiten:|Wunsch löschen:|Reservierung von)/})).toHaveCount(0);
  await page.reload(); await expect(page.locator('.claimant-name')).toHaveText('Auntie');
  const year=Number((await page.getByText(/Parent access|Elternzugang/).textContent()).match(/\d+/)[0]);
  const storage=`/data/households/SmokeFamily/${year}/christmas/Anna/wishlist.yaml`;
  const {mkdtempSync,readFileSync,rmSync}=require('node:fs'); const {tmpdir}=require('node:os'); const {join}=require('node:path');
  const dir=mkdtempSync(join(tmpdir(),'wishmas-volume-check-'));
  let saved;
  try { docker('cp',`${name}:${storage}`,join(dir,'wish.yaml'));saved=readFileSync(join(dir,'wish.yaml'),'utf8');assert.match(saved,/Wooden train/);assert.match(saved,/Auntie/); } finally {rmSync(dir,{recursive:true,force:true});}
  for(const path of ['/assets/app.css','/assets/app.js']) {const response=await fetch(origin+path);assert.equal(response.status,200);assert.ok((await response.text()).length>100);}
  docker('restart',name); origin=await ready();
  await page.goto(origin+'/'+parentPath); await page.getByRole('link',{name:'Anna',exact:true}).click();
  await expect(page.locator('.claimant-name')).toHaveText('Auntie');
  docker('rm','-f',name); origin=await start();
  await page.goto(origin+'/'+parentPath); await page.getByRole('link',{name:'Anna',exact:true}).click();
  await expect(page.locator('.claimant-name')).toHaveText('Auntie');
  await expect(page.getByText(text('Reference price: 49.90','Richtpreis: 49.90'),{exact:true})).toBeVisible();
  await guest.goto(origin+'/'+familyPath);await guest.getByRole('link',{name:'Anna',exact:true}).click();
  await expect(guest.getByText(text('Claimed by you','Von dir reserviert'),{exact:true})).toBeVisible();
  assert.equal(await guest.locator('.claimant-name').count(),0);
  assert.equal((await page.goto(origin)).status(),404);
  const after=mkdtempSync(join(tmpdir(),'wishmas-volume-check-'));
  try {docker('cp',`${name}:${storage}`,join(after,'wish.yaml'));assert.equal(readFileSync(join(after,'wish.yaml'),'utf8'),saved);}finally{rmSync(after,{recursive:true,force:true});}
  console.log(`PASS (${german ? 'de-CH' : 'en-US'}): container setup, embedded assets, roles, claims, named-volume YAML, restart and replacement.`);
 } finally {
  if(browser) await browser.close();
  for(const args of [['rm','-f',name],['volume','rm',volume],['image','rm',image]]) {try {docker(...args);} catch {}}
 }
})().catch(error=>{console.error(error);process.exitCode=1;});
