const fs = require('fs');
const { WASI } = require('wasi');

const inputUrl = process.argv[2] || 'https://example.com';

const wasi = new WASI({
  args: ['qrterminal.wasm', inputUrl],
  env: process.env,
  version: 'preview1',
});

const importObject = { wasi_snapshot_preview1: wasi.wasiImport };

(async () => {
  const wasm = await WebAssembly.compile(fs.readFileSync('./target/wasm32-wasip1/release/qrterminal.wasm'));
  const instance = await WebAssembly.instantiate(wasm, importObject);
  wasi.start(instance);
})();
