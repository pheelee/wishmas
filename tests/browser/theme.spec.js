const { test, expect } = require('./fixtures');
const { expectNightTheme } = require('./night-assertions');

// @test CON-003-AC1 CON-003-AC2 CON-003-AC3 CON-003-AC5
test('December night appearance and margin-only decorations fit desktop and mobile', async ({ page }) => {
  await page.goto('/');
  for (const [name, width, height] of [['desktop', 1536, 1024], ['mobile', 390, 844]]) {
    await page.setViewportSize({ width, height });
    await expectNightTheme(page);
    const bulbs = page.locator('[data-christmas-lights] .bulb');
    expect(await bulbs.count()).toBeGreaterThanOrEqual(4);
    const colors = await bulbs.evaluateAll(elements => elements.map(element => getComputedStyle(element).backgroundColor));
    expect(new Set(colors).size).toBe(4);
    expect(await bulbs.evaluateAll(elements => elements.every(element => getComputedStyle(element).animationName === 'none'))).toBe(true);
    await expect(page.getByRole('heading', { name: 'Make room for wishes.' })).toBeVisible();
    await expect(page.getByLabel('Household name')).toBeVisible();
    await expect(page.getByLabel('Household timezone')).toBeVisible();
    const flakes = await page.locator('[data-snow] .snowflake').evaluateAll(elements => elements.map(element => {
      const style = getComputedStyle(element);
      const rect = element.getBoundingClientRect();
      const main = document.querySelector('main').getBoundingClientRect();
      return {
        size: parseFloat(style.width),
        opacity: parseFloat(style.opacity),
        duration: parseFloat(style.animationDuration),
        outsideContent: rect.right <= main.left + 24 || rect.left >= main.right - 24,
        pointerEvents: style.pointerEvents,
      };
    }));
    expect(flakes.length).toBeGreaterThan(0);
    expect(flakes.length).toBeLessThanOrEqual(16);
    for (const flake of flakes) {
      expect(flake.size).toBeGreaterThanOrEqual(1);
      expect(flake.size).toBeLessThanOrEqual(3);
      expect(flake.opacity).toBeLessThanOrEqual(.35);
      expect(flake.duration).toBeGreaterThanOrEqual(20);
      expect(flake.duration).toBeLessThanOrEqual(40);
      expect(flake.outsideContent).toBe(true);
      expect(flake.pointerEvents).toBe('none');
    }
    await expect(page.locator('[data-snow]')).toHaveAttribute('aria-hidden', 'true');
    await expect(page.locator('[data-snow]')).toHaveAttribute('inert', '');
    await page.screenshot({ path: test.info().outputPath(`december-night-${name}.png`), fullPage: true });
  }
});

// @test CON-003-AC3 CON-003-AC4
test('snow always moves gently without pause controls and ignores the obsolete preference', async ({ page }) => {
  await page.goto('/');
  await page.evaluate(() => localStorage.setItem('wishmas.snowPaused', 'true'));
  await page.reload();
  const firstFlake = page.locator('[data-snow] .snowflake').first();
  await expect(page.getByRole('button', { name: /snow/i })).toHaveCount(0);
  await expect(firstFlake).toHaveCSS('animation-play-state', 'running');
  const initial = await firstFlake.evaluate(element => getComputedStyle(element).transform);
  await expect.poll(() => firstFlake.evaluate(element => getComputedStyle(element).transform)).not.toBe(initial);
  await page.emulateMedia({ reducedMotion: 'reduce' });
  await expect(firstFlake).toHaveCSS('animation-name', 'none');
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  await expect(firstFlake).toHaveCSS('animation-play-state', 'running');
});

// @test CON-003-AC4
test('reduced motion prevents snowfall from the first page load', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'reduce' });
  await page.goto('/');
  const flakes = page.locator('[data-snow] .snowflake');
  expect(await flakes.count()).toBeGreaterThan(0);
  expect(await flakes.evaluateAll(elements => elements.every(element => getComputedStyle(element).animationName === 'none'))).toBe(true);
  await expect(page.getByRole('button', { name: /snow/i })).toHaveCount(0);
  const rect = await flakes.first().boundingBox();
  expect(rect.y).toBeGreaterThan(0); expect(rect.y).toBeLessThan(844);
});

// @test CON-003-AC1 CON-003-AC5
test('validation, long household names, access pages, and not-found pages keep the theme', async ({ page }) => {
  await page.setViewportSize({ width: 320, height: 740 });
  await page.goto('/');
  await page.getByLabel('Household name').fill('Family!');
  await page.getByLabel('Household timezone').fill('Europe/Zurich');
  await page.getByRole('button', { name: 'Create access links' }).click();
  await expect(page.getByRole('alert')).toBeVisible();
  await expectNightTheme(page);
  await page.getByLabel('Household name').fill('É'.repeat(50));
  await page.getByRole('button', { name: 'Create access links' }).click();
  await expect(page.getByRole('button', { name: 'Finish setup' })).toBeVisible();
  await expectNightTheme(page);
  const family = await page.getByLabel('Family link', { exact: true }).inputValue();
  await page.getByRole('button', { name: 'Finish setup' }).click();
  await expect(page.getByText(/Parent access/)).toBeVisible();
  await expectNightTheme(page);
  await page.goto(family);
  await expect(page.getByText(/Family access/)).toBeVisible();
  await expectNightTheme(page);
  const missing = await page.goto('/?code=invalid');
  expect(missing.status()).toBe(404);
  await expect(page.getByText('404 page not found', { exact: true })).toBeVisible();
  await expectNightTheme(page);
});

// @test CON-003-AC2 CON-003-AC5 CON-003-AC6
test('bulb stems meet the cord at every size on desktop and mobile', async ({ page }) => {
 await page.goto('/');
 for (const width of [1440, 390]) {
  await page.setViewportSize({width,height:900});
  for (const size of ['80','120','200']) {
   await page.getByLabel('Bulb size').fill(size);
   await page.getByLabel('Bulb size').dispatchEvent('input');
   const distances = await page.evaluate(() => {
    const path = document.querySelector('.light-strand path'), matrix = path.getScreenCTM();
    return [...document.querySelectorAll('.bulb')].map(bulb => {
     const stem = getComputedStyle(bulb, '::before');
     const tip = document.createElement('i');
     // Measure the actual transformed center of the stem's top edge.
     Object.assign(tip.style,{position:'absolute',left:stem.left,top:`-${parseFloat(stem.height)}px`,width:stem.width,height:'0',transform:stem.transform});
     bulb.append(tip); const rect=tip.getBoundingClientRect(); tip.remove();
     const x=rect.x+rect.width/2, y=rect.y+rect.height/2;
     let low=0, high=path.getTotalLength();
     for(let i=0;i<30;i++){const mid=(low+high)/2;const p=path.getPointAtLength(mid).matrixTransform(matrix);if(p.x<x)low=mid;else high=mid;}
     const cord=path.getPointAtLength((low+high)/2).matrixTransform(matrix);
     return Math.abs(y-cord.y);
    });
   });
   expect(Math.max(...distances)).toBeLessThan(.75);
  }
 }
});
