const { test: base, expect } = require('@playwright/test');
const { spawn, execFile } = require('node:child_process');
const { mkdtempSync, rmSync } = require('node:fs');
const { tmpdir } = require('node:os');
const { join } = require('node:path');
const { createServer } = require('node:net');
const https = require('node:https');
const { promisify } = require('node:util');

// All browser checks exercise production New over isolated HTTPS.
const test = base.extend({
  appBinary: [async ({}, use) => {
    const dir = mkdtempSync(join(tmpdir(), 'wishmas-browser-build-'));
    const binary = join(dir, 'wishmas');
    try {
      await promisify(execFile)('go', ['test', '-c', '-o', binary, './internal/app']);
      await use(binary);
    } finally { rmSync(dir, { recursive: true, force: true }); }
  // @impl DEP-001-AC3 REL-001-AC2: cold Go/codec builds may exceed a page-test timeout.
  }, { scope: 'worker', timeout: 180000 }],
  appDataDir: async ({}, use) => {
    const dir = mkdtempSync(join(tmpdir(), 'wishmas-browser-data-'));
    try { await use(dir); } finally { rmSync(dir, { recursive: true, force: true }); }
  },
  appEnvironment: async ({}, use) => { await use({}); },
  appControl: async ({ appBinary, appDataDir, appEnvironment }, use) => {
    const port = await new Promise((resolve, reject) => {
      const socket = createServer();socket.once('error', reject);
      socket.listen(0, '127.0.0.1', () => {const number=socket.address().port;socket.close(error=>error?reject(error):resolve(number));});
    });
    const origin=`https://127.0.0.1:${port}`;
    let server, launchError, log='';
    async function start() {
      launchError=undefined;log='';
      server=spawn(appBinary,[],{env:{...process.env,WISHMAS_ADDR:`127.0.0.1:${port}`,WISHMAS_DATA_DIR:appDataDir,WISHMAS_BROWSER_TEST_MODE:'household',...appEnvironment},stdio:['ignore','ignore','pipe']});
      server.stderr.on('data',chunk=>{log+=chunk;});server.on('error',error=>{launchError=error;});
      const deadline=Date.now()+10000;
      while(true) {
        if(launchError)throw launchError;
        if(server.exitCode!==null||server.signalCode!==null)throw new Error(`Test server stopped: ${log}`);
        try {
          await new Promise((resolve,reject)=>{const request=https.get(origin,{rejectUnauthorized:false},response=>{response.resume();resolve();});request.once('error',reject);request.setTimeout(1000,()=>request.destroy(new Error('TLS fixture timeout')));});
          return;
        } catch(error) {
          if(Date.now()>deadline)throw new Error(`Test server did not start: ${log}`,{cause:error});
          await new Promise(resolve=>setTimeout(resolve,50));
        }
      }
    }
    async function stop() {
      if(server&&server.exitCode===null&&server.signalCode===null&&!launchError){const exited=new Promise(resolve=>server.once('exit',resolve));server.kill('SIGTERM');await exited;}
    }
    try {await start();await use({origin,restart:async()=>{await stop();await start();}});}finally{await stop();}
  },
  baseURL: async ({appControl},use)=>{await use(appControl.origin);},
});
module.exports={test,expect};
