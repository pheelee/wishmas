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
