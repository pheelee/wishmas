const {test,expect}=require('./fixtures');
const {expectNightTheme}=require('./night-assertions');
const {readFileSync,writeFileSync,mkdirSync}=require('node:fs');
const {join,dirname}=require('node:path');

// @test INT-001-AC1 INT-001-AC2 INT-001-AC3 INT-002-AC1 INT-002-AC2 INT-002-AC3 INT-002-AC4 INT-002-AC6
 test('setup keeps its browser language until Finish setup and translates browser validation',async({browser,baseURL})=>{
 const context=await browser.newContext({locale:'de-DE',baseURL}); const page=await context.newPage();
 await context.grantPermissions(['clipboard-read','clipboard-write']);
 await page.goto('/');
 await expect(page.locator('html')).toHaveAttribute('lang','de-CH');
 await expect(page.getByLabel('Sprache',{exact:true})).toHaveValue('de-CH');
 await page.getByLabel('Sprache',{exact:true}).selectOption('en-US');
 await page.getByRole('button',{name:'Zugangslinks erstellen'}).click();
 expect(await page.getByLabel('Name des Haushalts',{exact:true}).evaluate(e=>e.validationMessage)).toBe('Bitte fülle dieses Feld aus.');
 await page.getByLabel('Name des Haushalts',{exact:true}).fill('Family');
 await page.getByLabel('Zeitzone des Haushalts',{exact:true}).fill('Wrong/Zone');
 await page.getByLabel('Grösse der Lämpchen',{exact:true}).fill('180');
 await page.getByRole('button',{name:'Zugangslinks erstellen'}).click();
 await expect(page.getByRole('alert')).toContainText('Wähle eine gültige Zeitzone');
 await expect(page.getByLabel('Sprache',{exact:true})).toHaveValue('en-US');
 await expect(page.getByLabel('Grösse der Lämpchen',{exact:true})).toHaveValue('180');
 await page.getByLabel('Zeitzone des Haushalts',{exact:true}).fill('Europe/Zurich');
 await page.getByRole('button',{name:'Zugangslinks erstellen'}).click();
 await expect(page.getByText('Fast fertig',{exact:true})).toBeVisible();
 await page.getByRole('button',{name:'Elternlink kopieren'}).click();
 await expect(page.getByRole('status')).toHaveText('Link kopiert.');
 await page.getByRole('button',{name:'Einrichtung abschliessen'}).click();
 await expect(page.locator('html')).toHaveAttribute('lang','en-US');
 await expect(page.getByRole('heading',{name:'Christmas at Family.'})).toBeVisible();
 await expect(page.getByLabel('Language',{exact:true})).toHaveCount(0);
 await context.close();
 });

// @test INT-001-AC1 INT-001-AC2 INT-001-AC3 INT-001-AC4 INT-003-AC1 INT-003-AC4 INT-003-AC5 INT-003-AC6 INT-004-AC1 INT-004-AC2 INT-004-AC3 INT-005-AC1 INT-005-AC2 INT-005-AC3 INT-005-AC4 INT-005-AC5 CON-005
 test('German household preserves content, permissions, prices and sorting across manual language changes',async({page,context,browser,appDataDir})=>{
 test.setTimeout(60000);
 const requests=[];page.on('request',r=>requests.push(r.url()));
 await page.goto('/');await page.getByLabel('Household name').fill('Family');await page.getByLabel('Household timezone').fill('Europe/Zurich');
 await page.getByLabel('Language',{exact:true}).selectOption('de-CH');
 await page.getByRole('button',{name:'Create access links'}).click();
 const family=await page.getByLabel('Family link',{exact:true}).inputValue();
 await page.getByRole('button',{name:'Finish setup'}).click();const parent=page.url();
 await expect(page).toHaveTitle('Wishmas · Weihnachtswunschlisten');
 await page.getByLabel('Name der Person',{exact:true}).fill('Anna');await page.getByRole('button',{name:'Person hinzufügen',exact:true}).click();
 const year=Number((await page.getByText(/Elternzugang/).textContent()).match(/\d+/)[0]);
 await page.getByRole('link',{name:'Anna',exact:true}).click();const list=page.url();
 const add=page.getByRole('region',{name:'Wunsch hinzufügen',exact:true});
 await add.getByRole('button',{name:'Wunsch hinzufügen',exact:true}).click();
 expect(await add.getByLabel('Titel',{exact:true}).evaluate(e=>e.validationMessage)).toBe('Bitte fülle dieses Feld aus.');
 await add.getByLabel('Titel',{exact:true}).fill('Claim wish');await add.getByLabel('Richtpreis',{exact:true}).fill('1234.5');await add.getByLabel('Notiz',{exact:true}).fill('Your access links');await add.getByLabel('Shop-Links',{exact:true}).fill('https://shop.example/item');
 await add.getByRole('button',{name:'Wunsch hinzufügen',exact:true}).click();
 await expect(page.locator('.wish-title')).toHaveText('Claim wish');await expect(page.locator('.wish-note')).toHaveText('Your access links');
 await expect(page.locator('.wish-price')).toHaveText('Richtpreis: 1’234.50');
 await expect(page.getByRole('link',{name:'shop.example',exact:true})).toHaveAttribute('href','https://shop.example/item');
 await page.getByRole('button',{name:'Wunsch bearbeiten: Claim wish',exact:true}).click();
 const edit=page.getByRole('dialog',{name:'Wunsch bearbeiten',exact:true});
 await expect(edit.getByLabel('Richtpreis',{exact:true})).toHaveValue('1234.5');
 await edit.getByLabel('Richtpreis',{exact:true}).fill('1’234.50');await edit.getByRole('button',{name:'Wunsch speichern',exact:true}).click();
 const correction=page.getByRole('dialog',{name:'Angaben zum Wunsch korrigieren',exact:true});
 await expect(correction.getByRole('alert')).toContainText('Dezimalpunkt');await correction.getByRole('button',{name:'Abbrechen',exact:true}).click();await page.goto(list);
 const guestContext=await browser.newContext({locale:'en-US'}), guest=await guestContext.newPage();
 try {
 await guest.goto(family);await guest.getByRole('link',{name:'Anna',exact:true}).click();
 await expect(guest.locator('html')).toHaveAttribute('lang','de-CH');
 await guest.getByRole('button',{name:'Wunsch reservieren: Claim wish',exact:true}).click();
 await guest.getByLabel('Name für Reservierungen',{exact:true}).fill('Auntie');await guest.getByRole('button',{name:'Namen speichern und reservieren',exact:true}).click();
 await expect(guest.getByText('Von dir reserviert',{exact:true})).toBeVisible();
 await page.reload();await expect(page.locator('.claimant-name')).toHaveText('Auntie');
 await page.setViewportSize({width:1440,height:1000});await expectNightTheme(page);
 await page.screenshot({path:test.info().outputPath('german-desktop.png'),fullPage:true});
 await page.setViewportSize({width:390,height:844});await expectNightTheme(page);
 await page.screenshot({path:test.info().outputPath('german-mobile.png'),fullPage:true});
 await expect(guest.locator('.claimant-name')).toHaveCount(0);await expect(guest.getByRole('button',{name:/^(Wunsch bearbeiten:|Wunsch löschen:|Reservierung von)/})).toHaveCount(0);
 const file=join(appDataDir,'households','Family',String(year),'christmas','Anna','wishlist.yaml');const snapshot=readFileSync(file,'utf8');
 const metadata=join(appDataDir,'households','Family','household.yaml');const before=readFileSync(metadata,'utf8');
 await page.getByLabel('Wünsche sortieren',{exact:true}).selectOption('title');
 writeFileSync(metadata,before.replace('language: de-CH','language: en-US'));
 // Open pages remain German until a request. A form error uses the new language.
 await expect(page.locator('html')).toHaveAttribute('lang','de-CH');
 await page.getByRole('button',{name:'Wunsch bearbeiten: Claim wish',exact:true}).click();
 await edit.getByLabel('Titel',{exact:true}).fill('bad');await edit.getByRole('button',{name:'Wunsch speichern',exact:true}).click();
 await expect(page.locator('html')).toHaveAttribute('lang','en-US');
 await expect(page.getByRole('dialog',{name:'Correct wish details',exact:true}).getByRole('alert')).toHaveText('Use 5–150 characters for the title.');
 await page.getByRole('button',{name:'Cancel',exact:true}).filter({visible:true}).click();
 await expect(page.getByLabel('Sort wishes',{exact:true})).toHaveValue('title');
 await expect(page.locator('.wish-price')).toHaveText('Reference price: 1,234.50');
 expect(readFileSync(file,'utf8')).toBe(snapshot);
 await guest.reload();await expect(guest.getByText('Claimed by you',{exact:true})).toBeVisible();
 const old=join(appDataDir,'households','Family',String(year-1),'christmas','Anna','wishlist.yaml');mkdirSync(dirname(old),{recursive:true});writeFileSync(old,snapshot);
 const history=new URL(list);history.searchParams.set('year',String(year-1));await page.goto(history.toString());
 await expect(page.getByText('Read-only history',{exact:true})).toBeVisible();await expect(page.getByRole('button',{name:/^(Edit wish:|Delete wish:|Claim wish:|Remove my claim:)/})).toHaveCount(0);
 writeFileSync(metadata,before);await page.reload();await expect(page.getByText('Vergangene Jahre – nur lesbar',{exact:true})).toBeVisible();
 const missing=new URL(parent);missing.pathname='/unknown';await page.goto(missing.toString());await expect(page.getByRole('heading')).toHaveText('404 – Seite nicht gefunden');
 await page.goto(parent);writeFileSync(join(appDataDir,'households','Family','people.yaml'),'people: [broken]\n');await page.reload();await expect(page.getByRole('heading')).toContainText('Die Wunschlistendaten konnten nicht gelesen werden.');
 await page.goto('/?code=invalid');await expect(page.locator('html')).toHaveAttribute('lang','en-US');
 expect(requests.every(url=>new URL(url).origin===new URL(parent).origin)).toBe(true);
 } finally {await guestContext.close();}
 });
