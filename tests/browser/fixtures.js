const { test: base, expect } = require('@playwright/test');
const { spawn, execFile } = require('node:child_process');
const { mkdtempSync, rmSync } = require('node:fs');
const { tmpdir } = require('node:os');
const { join } = require('node:path');
const { createServer } = require('node:net');
const { promisify } = require('node:util');

// Each test gets its own process and household so setup-dependent tests are independent.
const test = base.extend({
  appBinary: [async ({}, use) => {
    const dir = mkdtempSync(join(tmpdir(), 'wishmas-browser-build-'));
    const binary = join(dir, 'wishmas');
    try {
      await promisify(execFile)('go', ['build', '-o', binary, './cmd/wishmas']);
      await use(binary);
    } finally {
      rmSync(dir, { recursive: true, force: true });
    }
  }, { scope: 'worker' }],
  baseURL: async ({ appBinary }, use) => {
    const port = await new Promise((resolve, reject) => {
      const socket = createServer();
      socket.once('error', reject);
      socket.listen(0, '127.0.0.1', () => {
        const number = socket.address().port;
        socket.close(error => error ? reject(error) : resolve(number));
      });
    });
    const dir = mkdtempSync(join(tmpdir(), 'wishmas-browser-data-'));
    const origin = `http://127.0.0.1:${port}`;
    const server = spawn(appBinary, [], {
      env: { ...process.env, WISHMAS_ADDR: `127.0.0.1:${port}`, WISHMAS_DATA_DIR: dir },
      stdio: ['ignore', 'ignore', 'pipe'],
    });
    let log = '';
    let launchError;
    server.stderr.on('data', chunk => { log += chunk; });
    server.on('error', error => { launchError = error; });
    try {
      const deadline = Date.now() + 10000;
      while (true) {
        if (launchError) throw launchError;
        if (server.exitCode !== null) throw new Error(`Test server stopped: ${log}`);
        try {
          await fetch(origin, { signal: AbortSignal.timeout(1000) });
          break;
        } catch (error) {
          if (Date.now() > deadline) throw new Error(`Test server did not start: ${log}`, { cause: error });
          await new Promise(resolve => setTimeout(resolve, 50));
        }
      }
      await use(origin);
    } finally {
      if (server.exitCode === null && !launchError) {
        const exited = new Promise(resolve => server.once('exit', resolve));
        server.kill('SIGTERM');
        await exited;
      }
      rmSync(dir, { recursive: true, force: true });
    }
  },
});

module.exports = { test, expect };
