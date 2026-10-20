// @impl ACC-001-AC2
const timezones = document.querySelector('#timezones');
if (timezones && typeof Intl.supportedValuesOf === 'function') {
  for (const zone of Intl.supportedValuesOf('timeZone')) {
    const option = document.createElement('option');
    option.value = zone;
    timezones.append(option);
  }
}

// @impl ACC-001-AC6 ACC-003-AC2
for (const id of ['parent-link', 'family-link']) {
  const input = document.getElementById(id);
  if (input) input.value = new URL(input.value, window.location.origin).href;
}
for (const button of document.querySelectorAll('[data-copy]')) {
  button.addEventListener('click', async () => {
    const input = document.getElementById(button.dataset.copy);
    const status = document.getElementById('copy-status');
    try {
      if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(input.value);
      } else {
        input.select();
        if (!document.execCommand('copy')) throw new Error('copy failed');
      }
      status.textContent = 'Link copied.';
    } catch {
      input.focus();
      input.select();
      status.textContent = 'Select and copy the link above.';
    }
  });
}

// @impl CON-003-AC3 CON-003-AC4
const snow = document.querySelector('[data-snow]');
const snowToggle = document.querySelector('[data-snow-toggle]');
if (snow && snowToggle) {
  const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)');
  let paused = false;
  try { paused = localStorage.getItem('wishmas.snowPaused') === 'true'; } catch { /* Keep the control usable if storage is unavailable. */ }
  const updateSnow = () => {
    const effectivePause = paused || reducedMotion.matches;
    snow.dataset.paused = String(effectivePause);
    snowToggle.textContent = effectivePause ? 'Resume snow' : 'Pause snow';
    snowToggle.disabled = reducedMotion.matches;
    snowToggle.title = reducedMotion.matches ? 'Snow is paused by your reduced motion preference.' : '';
  };
  snowToggle.hidden = false;
  updateSnow();
  reducedMotion.addEventListener('change', updateSnow);
  snowToggle.addEventListener('click', () => {
    paused = !paused;
    try { localStorage.setItem('wishmas.snowPaused', String(paused)); } catch { /* Pause still works for this page. */ }
    updateSnow();
  });
}

// @impl WSH-003-AC1 WSH-003-AC2 WSH-003-AC3 WSH-003-AC4 WSH-003-AC5 WSH-003-AC6 WSH-003-AC7 WSH-003-AC8
const wishList = document.querySelector('[data-wish-list]');
const wishSort = document.querySelector('[data-wish-sort]');
if (wishList && wishSort) {
  const preferenceKey = 'wishmas.wishSort';
  const orders = new Set(['oldest', 'newest', 'title', 'price-asc', 'price-desc']);
  // Preserve the saved sequence independently of the current display order.
  // Move whole cards so their forms and wish tokens remain attached.
  const cards = [...wishList.children].map((element, added) => {
    const [whole, fraction = ''] = (element.dataset.wishPrice || '0').split('.');
    return {
      element, added, title: element.dataset.wishTitle,
      // Integer hundredths avoid rounding even for large reference prices.
      price: BigInt(whole) * 100n + BigInt(fraction.padEnd(2, '0')),
    };
  });
  const titles = new Intl.Collator('en', { sensitivity: 'base' });
  const applySort = () => {
    const order = wishSort.value;
    const sorted = [...cards].sort((a, b) => {
      let comparison = 0;
      if (order === 'newest') comparison = b.added - a.added;
      if (order === 'title') comparison = titles.compare(a.title, b.title);
      if (order === 'price-asc' || order === 'price-desc') {
        comparison = a.price < b.price ? -1 : a.price > b.price ? 1 : 0;
        if (order === 'price-desc') comparison *= -1;
      }
      return comparison || a.added - b.added;
    });
    wishList.append(...sorted.map(card => card.element));
  };
  let preference;
  try { preference = localStorage.getItem(preferenceKey); } catch { /* Sorting still works without storage. */ }
  wishSort.value = orders.has(preference) ? preference : 'oldest';
  applySort();
  document.querySelector('[data-wish-sort-control]').hidden = false;
  wishSort.addEventListener('change', () => {
    applySort();
    try { localStorage.setItem(preferenceKey, wishSort.value); } catch { /* Keep the selected order for this page. */ }
  });
}
