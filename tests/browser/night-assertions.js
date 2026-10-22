const { expect } = require('@playwright/test');

async function expectNightTheme(page) {
  // Native form submissions can commit before the new stylesheet has loaded.
  await page.waitForLoadState("load");
  const colors = await page.evaluate(() => {
    const canvas = document.createElement('canvas');
    canvas.width = canvas.height = 1;
    const context = canvas.getContext('2d', { willReadFrequently: true });
    const rgb = value => {
      context.clearRect(0, 0, 1, 1);
      context.fillStyle = value;
      context.fillRect(0, 0, 1, 1);
      return Array.from(context.getImageData(0, 0, 1, 1).data).slice(0, 3);
    };
    const luminance = color => rgb(color).map(c => {
      c /= 255;
      return c <= .04045 ? c / 12.92 : ((c + .055) / 1.055) ** 2.4;
    }).reduce((sum, c, i) => sum + c * [.2126, .7152, .0722][i], 0);
    const background = getComputedStyle(document.body).backgroundColor;
    const inputs = [...document.querySelectorAll('main input:not([type="hidden"]), main textarea, main select')];
    const panels = [...document.querySelectorAll('main .panel, main .wish-card, dialog.modal')];
    const surfaces = [...panels, ...inputs];
    const labels = [...document.querySelectorAll('main label')];
    return {
      background: luminance(background),
      surfaces: surfaces.map(element => luminance(getComputedStyle(element).backgroundColor)),
      labelContrasts: labels.map(label => {
        const fg = luminance(getComputedStyle(label).color);
        const bg = luminance(getComputedStyle(label.closest('.panel') || label.closest('form') || document.body).backgroundColor);
        return (Math.max(fg, bg) + .05) / (Math.min(fg, bg) + .05);
      }),
      fits: document.documentElement.scrollWidth <= window.innerWidth,
    };
  });
  expect(colors.background).toBeLessThan(.03);
  expect(colors.surfaces.length).toBeGreaterThan(0);
  for (const luminance of colors.surfaces) expect(luminance).toBeLessThan(.08);
  for (const contrast of colors.labelContrasts) expect(contrast).toBeGreaterThanOrEqual(4.5);
  expect(colors.fits).toBe(true);
}

module.exports = { expectNightTheme };
