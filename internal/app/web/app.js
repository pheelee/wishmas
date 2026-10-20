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
      button.dataset.copied = 'true';
      clearTimeout(button.copyReset);
      button.copyReset = setTimeout(() => { button.dataset.copied = 'false'; }, 3000);
    } catch {
      input.focus();
      input.select();
      status.textContent = 'Select and copy the link above.';
    }
  });
}

// @impl ACC-004-AC1
const timezoneInput = document.querySelector('#timezone');
if (timezoneInput && !timezoneInput.value) {
  try { timezoneInput.value = Intl.DateTimeFormat().resolvedOptions().timeZone || ''; } catch { /* Manual entry remains available. */ }
}

// @impl CON-003-AC3 CON-003-AC4 CON-003-AC6 CON-003-AC7 CON-003-AC8
function previewAtmosphere() {
  const body = document.body;
  const size = Number(body.dataset.bulbSize);
  body.style.setProperty('--bulb-size', size);
  const strand = document.querySelector('[data-christmas-lights]');
  strand.querySelectorAll('.bulb').forEach(bulb => bulb.remove());
  const count = Math.round(1800 / size);
  for (let i = 0; i < count; i++) {
    const bulb = document.createElement('span');
    const left = 2.5 + 95 * i / (count - 1), x = left / 100;
    bulb.className = `bulb bulb-${['coral', 'mint', 'gold', 'blue'][i % 4]}`;
    bulb.style.setProperty('--left', `${left}%`);
    bulb.style.setProperty('--drop', `${5 + 240 * x * (1 - x)}px`);
    strand.append(bulb);
  }
  const amount = Number(body.dataset.snowAmount), scale = Number(body.dataset.snowSize) / 100;
  document.querySelectorAll('.snow-side').forEach(side => side.replaceChildren());
  for (let i = 0; i < amount; i++) {
    const flake = document.createElement('span');
    flake.className = 'snowflake';
    const values = {x: `${5 + 90 * Math.floor(i / 2) / (Math.ceil(amount / 2) - 1)}%`, size: `${(1 + i % 3) * scale}px`, 'rest-top': `${5 + i * 17 % 90}%`, duration: `${20 + (i * 7 + 2) % 21}s`, delay: `${-(i * 11 % 40)}s`, opacity: .14 + i % 5 * .04};
    for (const [key, value] of Object.entries(values)) flake.style.setProperty(`--${key}`, value);
    document.querySelector(i % 2 === 0 ? '.snow-side-left' : '.snow-side-right').append(flake);
  }
}
for (const slider of document.querySelectorAll('[data-appearance]')) {
  slider.addEventListener('input', () => {
    document.body.dataset[slider.dataset.appearance] = slider.value;
    document.querySelector(`[data-range-output="${slider.id}"]`).value = slider.value + (slider.id === 'snow-amount' ? '' : '%');
    previewAtmosphere();
  });
}

// @impl CON-003-AC10 WSH-004-AC3 WSH-004-AC4 WLS-005-AC1 ACC-004-AC4
for (const trigger of document.querySelectorAll('[data-open-modal]')) {
  trigger.addEventListener('click', () => document.getElementById(trigger.dataset.openModal).showModal());
}
for (const button of document.querySelectorAll('[data-close-modal]')) {
  button.addEventListener('click', () => button.closest('dialog').close());
}
for (const modal of document.querySelectorAll('dialog[data-auto-open="true"]')) {
  if (modal.open) modal.close();
  modal.showModal();
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
