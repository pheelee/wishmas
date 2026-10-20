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
      status.textContent = document.body.dataset.copySuccess;
      button.dataset.copied = 'true';
      clearTimeout(button.copyReset);
      button.copyReset = setTimeout(() => { button.dataset.copied = 'false'; }, 3000);
    } catch {
      input.focus();
      input.select();
      status.textContent = document.body.dataset.copyFailure;
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

// @impl WSH-006-AC10: picture transitions respect reduced-motion preferences.
const pictureMotion = () => !window.matchMedia('(prefers-reduced-motion: reduce)').matches;
function openModal(modal) {
  modal.showModal();
  if (modal.classList.contains('picture-modal') && pictureMotion()) {
    modal.animate([{ opacity: 0, transform: 'scale(.97)' }, { opacity: 1, transform: 'scale(1)' }],
      { duration: 180, easing: 'ease-in-out' });
  }
}
function closeModal(modal) {
  if (!modal.classList.contains('picture-modal') || !pictureMotion()) {
    modal.close();
    return;
  }
  if (modal.classList.contains('is-closing')) return;
  const current = getComputedStyle(modal);
  const from = { opacity: current.opacity, transform: current.transform };
  for (const animation of modal.getAnimations()) animation.cancel();
  modal.classList.add('is-closing');
  const animation = modal.animate([from, { opacity: 0, transform: 'scale(.97)' }],
    { duration: 180, easing: 'ease-in-out', fill: 'forwards' });
  animation.finished.then(() => {
    modal.close();
    animation.cancel();
    modal.classList.remove('is-closing');
  });
}

// @impl CON-003-AC10 WSH-004-AC3 WSH-004-AC4 WLS-005-AC1 ACC-004-AC4
for (const trigger of document.querySelectorAll('[data-open-modal]')) {
  trigger.addEventListener('click', () => openModal(document.getElementById(trigger.dataset.openModal)));
}
// @impl WSH-006-AC10: only clicks outside the picture dialog dismiss it.
for (const modal of document.querySelectorAll('.picture-modal')) {
  modal.addEventListener('cancel', event => { event.preventDefault(); closeModal(modal); });
  modal.addEventListener('click', event => {
    if (event.target !== modal) return;
    const bounds = modal.getBoundingClientRect();
    if (event.clientX < bounds.left || event.clientX > bounds.right ||
        event.clientY < bounds.top || event.clientY > bounds.bottom) closeModal(modal);
  });
}

for (const button of document.querySelectorAll('[data-close-modal]')) {
  button.addEventListener('click', () => closeModal(button.closest('dialog')));
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
  // @impl INT-005-AC5
  const titles = new Intl.Collator(document.documentElement.lang, { sensitivity: 'base' });
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

// @impl INT-001-AC3: override browser-owned required-field text using the active page language.
for (const field of document.querySelectorAll('input[required], textarea[required], select[required]')) {
 if (field.form) field.form.noValidate = false;
 const validate = () => {
  field.setCustomValidity('');
  if (field.validity.valueMissing) field.setCustomValidity(document.body.dataset.requiredMessage);
 };
 field.addEventListener('invalid', validate);
 field.addEventListener('input', () => field.setCustomValidity(''));
 field.addEventListener('change', () => field.setCustomValidity(''));
}

// @impl WSH-006-AC15: failed local images never trigger a remote retry.
for (const image of document.querySelectorAll('[data-picture-image]')) {
 const unavailable = () => {
  const container = image.closest('[data-picture-shelf], dialog');
  const button = image.closest('button');
  if (button) button.hidden = true; else image.hidden = true;
  container.querySelector('.picture-unavailable').hidden = false;
 };
 image.addEventListener('error', unavailable);
 if (image.complete && image.naturalWidth === 0) unavailable();
}
