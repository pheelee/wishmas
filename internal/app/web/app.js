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
