const {test,expect}=require('./fixtures');
const {expectNightTheme}=require('./night-assertions');
const {readFileSync,writeFileSync}=require('node:fs');
const {join}=require('node:path');
// TLS trust bypass is confined to the self-signed local test fixture.
test.use({appMode:'household',ignoreHTTPSErrors:true});
const password='PublicBrowserFixture1!';

// @test ACC-005 ACC-006-AC1 ACC-006-AC2 ACC-006-AC4 ACC-006-AC5 ACC-006-AC9 INT-006 INT-007 CON-007-AC4
for(const locale of ['en-US','de-CH'])test(`Password setup and persistent admin sessions in ${locale}`,async({browser,baseURL,appControl,appDataDir})=>{
 test.setTimeout(60000);
 const de=locale==='de-CH',label=(en,german)=>de?german:en;
 // Deliberately choose the opposite browser/setup display language.
 const setupDe=!de,setup=(en,german)=>setupDe?german:en;
 const context=await browser.newContext({baseURL,ignoreHTTPSErrors:true,locale:setupDe?'de-CH':'en-US'});const page=await context.newPage();
 await page.goto('/');await expect(page.locator('html')).toHaveAttribute('lang',setupDe?'de-CH':'en-US');
 await expect(page.locator('[name=timezone]')).toHaveCount(0);
 await page.getByLabel(setup('Household name','Name des Haushalts'),{exact:true}).fill('Family');
 await page.getByLabel(setup('Language','Sprache'),{exact:true}).selectOption(locale);
 for(const [id,value] of [['bulb-size','150'],['snow-size','80'],['snow-amount','24']])await page.locator('#'+id).fill(value);
 await expect(page.locator('body')).toHaveAttribute('data-bulb-size','150');
 await page.setViewportSize({width:1440,height:1000});await expectNightTheme(page);
 await page.screenshot({path:test.info().outputPath(`setup-${locale}-desktop.png`),fullPage:true});
 await page.getByLabel(setup('Admin password','Admin-Passwort'),{exact:true}).fill(password);
 await page.getByLabel(setup('Confirm password','Passwort bestätigen'),{exact:true}).fill('DifferentPublicFixture2!');
 await page.getByRole('button',{name:setup('Finish setup','Einrichtung abschliessen'),exact:true}).click();
 await expect(page.getByRole('alert')).toContainText(setup('The passwords do not match.','Die Passwörter stimmen nicht überein.'));
 for(const name of ['password','password_confirm'])await expect(page.locator(`[name=${name}]`)).toHaveValue('');
 await expect(page.locator('[name=household]')).toHaveValue('Family');await expect(page.locator('#language')).toHaveValue(locale);
 await expect(page.locator('#snow-amount')).toHaveValue('24');await expect(page.locator('html')).toHaveAttribute('lang',setupDe?'de-CH':'en-US');
 await page.getByLabel(setup('Admin password','Admin-Passwort'),{exact:true}).fill(password);
 await page.getByLabel(setup('Confirm password','Passwort bestätigen'),{exact:true}).fill(password);
 await page.getByRole('button',{name:setup('Finish setup','Einrichtung abschliessen'),exact:true}).click();
 await expect(page.getByText(label('Admin access','Admin-Zugang'),{exact:true})).toBeVisible();
 await expect(page.locator('html')).toHaveAttribute('lang',locale);
 await expect(page.getByText(label('No lists yet.','Noch keine Listen vorhanden.'),{exact:true})).toBeVisible();
 const cfg=join(appDataDir,'households','Family','household.yaml'),raw=readFileSync(cfg,'utf8');
 const cookie=(await context.cookies()).find(c=>c.name==='__Host-wishmas-admin');expect(!!cookie).toBe(true);
 expect(cookie.secure).toBe(true);expect(cookie.httpOnly).toBe(true);expect(cookie.sameSite).toBe('Lax');expect(cookie.expires).toBeGreaterThan(Date.now()/1000);
 expect(raw.includes(password)).toBe(false);expect(raw.includes(cookie.value)).toBe(false);expect(raw.includes('parent_code:')).toBe(false);
 // A fresh browser context restored from persistent browser storage stays signed in.
 const restored=await browser.newContext({baseURL,ignoreHTTPSErrors:true,storageState:await context.storageState()});
 const restoredPage=await restored.newPage();await restoredPage.goto('/');await expect(restoredPage.getByText(label('Admin access','Admin-Zugang'),{exact:true})).toBeVisible();await restored.close();
 // A real server process replacement retains the same session and data volume.
 await appControl.restart();await page.reload();await expect(page.getByText(label('Admin access','Admin-Zugang'),{exact:true})).toBeVisible();
 const second=await browser.newContext({baseURL,ignoreHTTPSErrors:true,locale:'fr-FR'}),guest=await second.newPage();
 await guest.goto('/?code=invalid');await expect(guest.getByRole('heading',{name:label('Admin sign-in','Admin-Anmeldung'),exact:true})).toBeVisible();await expect(guest.locator('html')).toHaveAttribute('lang',locale);
 await guest.getByLabel(label('Admin password','Admin-Passwort'),{exact:true}).fill(password);
 await guest.getByRole('button',{name:label('Sign in','Anmelden'),exact:true}).click();await expect(guest.getByText(label('Admin access','Admin-Zugang'),{exact:true})).toBeVisible();
 await guest.getByRole('button',{name:label('Log out','Abmelden'),exact:true}).click();await expect(guest.getByRole('heading',{name:label('Admin sign-in','Admin-Anmeldung'),exact:true})).toBeVisible();
 await page.reload();await expect(page.getByText(label('Admin access','Admin-Zugang'),{exact:true})).toBeVisible();
 await page.setViewportSize({width:390,height:844});await expectNightTheme(page);await page.screenshot({path:test.info().outputPath(`admin-${locale}-mobile.png`),fullPage:true});
 await guest.setViewportSize({width:390,height:844});await expectNightTheme(guest);await guest.screenshot({path:test.info().outputPath(`login-${locale}-mobile.png`),fullPage:true});
 await second.close();await context.close();
});

// @test ACC-006-AC6 CON-007-AC3 CON-007-AC4
 test('Cross-origin mutations fail and manual password changes revoke browser sessions',async({page,context,browser,baseURL,appDataDir})=>{
 await page.goto('/');await page.getByLabel('Household name',{exact:true}).fill('Family');
 await page.getByLabel('Language',{exact:true}).selectOption('en-US');
 await page.getByLabel('Admin password',{exact:true}).fill(password);await page.getByLabel('Confirm password',{exact:true}).fill(password);
 await page.getByRole('button',{name:'Finish setup',exact:true}).click();await expect(page.getByText('Admin access',{exact:true})).toBeVisible();
 const csrf=(await context.cookies()).find(c=>c.name==='__Host-wishmas-csrf').value;
 const rejected=await context.request.post('/logout',{form:{csrf},headers:{Origin:'https://attacker.example'}});expect(rejected.status()).toBe(403);
 await page.reload();await expect(page.getByText('Admin access',{exact:true})).toBeVisible();
 const second=await browser.newContext({baseURL,ignoreHTTPSErrors:true}),other=await second.newPage();await other.goto('/');await other.getByLabel('Admin password',{exact:true}).fill(password);await other.getByRole('button',{name:'Sign in',exact:true}).click();await expect(other.getByText('Admin access',{exact:true})).toBeVisible();
 // Cross-site requests never receive the private code-bearing page URL.
 await page.goto('/?code=private-referrer-fixture');let externalReferrer;
 await page.route('https://shop.example/privacy-check',async route=>{externalReferrer=(await route.request().allHeaders()).referer;await route.fulfill({status:200,body:''});});
 await page.evaluate(()=>fetch('https://shop.example/privacy-check',{mode:'no-cors'}));expect(externalReferrer).toBeUndefined();
 await page.goto('/');await expect(page.getByText('Admin access',{exact:true})).toBeVisible();
 const cfg=join(appDataDir,'households','Family','household.yaml'),raw=readFileSync(cfg,'utf8'),old=raw.match(/^admin_password_hash: (.+)$/m)[1],cut=old.lastIndexOf('$');
 const edited=old.slice(0,cut+1)+(old[cut+1]==='A'?'B':'A')+old.slice(cut+2);writeFileSync(cfg,raw.replace(old,edited));
 await page.reload();await expect(page.getByRole('heading',{name:'Admin sign-in',exact:true})).toBeVisible();
 await other.reload();await expect(other.getByRole('heading',{name:'Admin sign-in',exact:true})).toBeVisible();
 expect(/sessions: \[\]/.test(readFileSync(cfg,'utf8'))).toBe(true);
 await second.close();
});

// @test ACC-006-AC7 ACC-006-AC9
 test('German login blocks a sixth attempt and restart resets the block',async({page,appControl})=>{
 await page.goto('/');await page.getByLabel('Household name',{exact:true}).fill('Family');await page.getByLabel('Language',{exact:true}).selectOption('de-CH');
 await page.getByLabel('Admin password',{exact:true}).fill(password);await page.getByLabel('Confirm password',{exact:true}).fill(password);await page.getByRole('button',{name:'Finish setup',exact:true}).click();
 await page.getByRole('button',{name:'Abmelden',exact:true}).click();
 for(let i=0;i<5;i++) {await page.getByLabel('Admin-Passwort',{exact:true}).fill('WrongPublicFixture1!');const [response]=await Promise.all([page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith('/login')),page.getByRole('button',{name:'Anmelden',exact:true}).click()]);expect(response.status()).toBe(401);await expect(page.getByRole('alert')).toHaveText('Falsches Passwort.');}
 await page.getByLabel('Admin-Passwort',{exact:true}).fill(password);const [blocked]=await Promise.all([page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith('/login')),page.getByRole('button',{name:'Anmelden',exact:true}).click()]);expect(blocked.status()).toBe(429);
 await expect(page.getByRole('alert')).toContainText('15-minütigen Sperre');
 await appControl.restart();await page.goto('/');await page.getByLabel('Admin-Passwort',{exact:true}).fill(password);await page.getByRole('button',{name:'Anmelden',exact:true}).click();await expect(page.getByText('Admin-Zugang',{exact:true})).toBeVisible();
});
