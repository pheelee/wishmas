const { defineConfig } = require('@playwright/test');
const { mkdtempSync } = require('node:fs');
const { tmpdir } = require('node:os');
const { join } = require('node:path');

module.exports = defineConfig({
  testDir: './tests/browser',
  workers: 1,
  use: { baseURL: 'http://127.0.0.1:18080', browserName: 'chromium', screenshot: 'only-on-failure' },
  webServer: {
    command: 'go run ./cmd/wishmas',
    url: 'http://127.0.0.1:18080',
    reuseExistingServer: false,
    env: { WISHMAS_ADDR: '127.0.0.1:18080', WISHMAS_DATA_DIR: mkdtempSync(join(tmpdir(), 'wishmas-e2e-')) },
  },
});
