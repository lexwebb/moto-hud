import { copyFileSync, existsSync, mkdirSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const siteRoot = join(__dirname, '..');
const repoRoot = join(siteRoot, '..');

const encSrcDir = join(repoRoot, 'enclosure', 'exports');
const glbDstDir = join(siteRoot, 'public', 'enclosure');
const glbSrc = join(encSrcDir, 'assembly.glb');
const glbDst = join(glbDstDir, 'assembly.glb');

const routeSrc = join(repoRoot, 'web', 'emulator', 'routes', 'whitehall-farringdon.json');
const roadsSrc = join(repoRoot, 'web', 'emulator', 'routes', 'whitehall-farringdon-roads.json');
const routeDstDir = join(siteRoot, 'public', 'emulator', 'routes');
const routeDst = join(routeDstDir, 'whitehall-farringdon.json');
const roadsDst = join(routeDstDir, 'whitehall-farringdon-roads.json');

const execSrc = join(repoRoot, 'web', 'emulator', 'wasm_exec.js');
const execDst = join(siteRoot, 'public', 'emulator', 'wasm_exec.js');

const wasmSrc = join(repoRoot, 'web', 'emulator', 'motohud.wasm');
const wasmDst = join(siteRoot, 'public', 'emulator', 'motohud.wasm');

mkdirSync(glbDstDir, { recursive: true });
mkdirSync(routeDstDir, { recursive: true });
mkdirSync(dirname(execDst), { recursive: true });

if (existsSync(glbSrc)) {
  copyFileSync(glbSrc, glbDst);
  console.log('copied', glbDst);
} else {
  console.warn('missing', glbSrc);
}

for (const name of [
  'dock_interface.stl',
  'dock_plate_slab.stl',
  'dock_pod_slab.stl',
  'dock_plate_usb.stl',
]) {
  const src = join(encSrcDir, name);
  if (existsSync(src)) {
    copyFileSync(src, join(glbDstDir, name));
    console.log('copied', join(glbDstDir, name));
  } else {
    console.warn('missing', src);
  }
}

if (existsSync(routeSrc)) {
  copyFileSync(routeSrc, routeDst);
  console.log('copied', routeDst);
}

if (existsSync(roadsSrc)) {
  copyFileSync(roadsSrc, roadsDst);
  console.log('copied', roadsDst);
}

if (existsSync(execSrc)) {
  copyFileSync(execSrc, execDst);
  console.log('copied', execDst);
}

if (existsSync(wasmSrc)) {
  copyFileSync(wasmSrc, wasmDst);
  console.log('copied', wasmDst);
} else {
  console.warn(
    'motohud.wasm not found — run npm run build:wasm before deploy (CI does this). Emulator page will error until then.',
  );
}
