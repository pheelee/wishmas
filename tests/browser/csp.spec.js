const {test,expect}=require('./fixtures');
test.use({ignoreHTTPSErrors:true});

// @test OPS-001-AC8: execute eval from a real allowed bundle, not DevTools,
// whose privileged evaluation can bypass CSP and give misleading test results.
test('CSP allows the app bundle but blocks injected scripts and eval',async({page})=>{
 await page.route('**/assets/app.js',async route=>{
  const response=await route.fetch();
  const probe=`
window.cspProbe={loaded:true,evalBlocked:false,violations:[]};
document.addEventListener('securitypolicyviolation',e=>window.cspProbe.violations.push(e.effectiveDirective));
try { eval('window.cspEvalRan=true'); } catch(error) { window.cspProbe.evalBlocked=error instanceof EvalError; }
`;
  await route.fulfill({response,body:await response.text()+probe});
 });
 await page.route('https://csp-fixture.invalid/**',route=>route.fulfill({contentType:'application/javascript',body:'window.cspExternalRan=true;'}));
 await page.goto('/');
 await expect.poll(()=>page.evaluate(()=>window.cspProbe?.loaded)).toBe(true);
 expect(await page.evaluate(()=>window.cspProbe.evalBlocked)).toBe(true);
 await page.evaluate(()=>{
  const script=document.createElement('script');script.textContent='window.cspInlineRan=true';document.head.append(script);
  const button=document.createElement('button');button.setAttribute('onclick','window.cspHandlerRan=true');document.body.append(button);button.click();
  const external=document.createElement('script');external.src='https://csp-fixture.invalid/script.js';document.head.append(external);
 });
 await expect.poll(()=>page.evaluate(()=>['script-src','script-src-elem','script-src-attr'].every(d=>window.cspProbe.violations.includes(d)))).toBe(true);
 expect(await page.evaluate(()=>[window.cspEvalRan,window.cspInlineRan,window.cspHandlerRan,window.cspExternalRan].some(Boolean))).toBe(false);
 await expect.poll(()=>page.locator('[data-top-snow] .snow-top-flake').count()).toBeGreaterThan(0);
});

// @test OPS-001-AC8: no outside endpoint is contacted even if a policy regresses.
test('CSP rejects base changes, embedded objects and external form submissions',async({page})=>{
 const outsideRequests=[];
 await page.route('https://csp-fixture.invalid/**',route=>{outsideRequests.push(route.request().url());return route.abort();});
 await page.goto('/');
 const original=page.url();
 await page.evaluate(()=>{
  window.cspViolations=[];
  document.addEventListener('securitypolicyviolation',e=>window.cspViolations.push(e.effectiveDirective));
  const base=document.createElement('base');base.href='https://csp-fixture.invalid/base/';document.head.append(base);
  const object=document.createElement('object');object.data='https://csp-fixture.invalid/object';document.body.append(object);
  const form=document.createElement('form');form.method='POST';form.action='https://csp-fixture.invalid/submit';document.body.append(form);form.submit();
 });
 await expect.poll(()=>page.evaluate(()=>['base-uri','object-src','form-action'].every(d=>window.cspViolations.includes(d)))).toBe(true);
 expect(await page.evaluate(()=>document.baseURI)).toBe(original);
 expect(page.url()).toBe(original);
 expect(outsideRequests).toEqual([]);
});
