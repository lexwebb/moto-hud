import { copyFileSync, existsSync, mkdirSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';

const __dirname = dirname(fileURLToPath(import.meta.url));
const siteRoot = join(__dirname, '..');
const repoRoot = join(siteRoot, '..');
const outDir = join(siteRoot, 'public', 'emulator');
const outWasm = join(outDir, 'motohud.wasm');
const outExec = join(outDir, 'wasm_exec.js');

mkdirSync(outDir, { recursive: true });

const isWin = process.platform === 'win32';
const script = join(repoRoot, 'scripts', isWin ? 'build-wasm.sh' : 'build-wasm.sh');

function runGoWasmBuild() {
  // Prefer the repo script under bash when available; else inline go build.
  const bash = spawnSync('bash', [script], { cwd: repoRoot, encoding: 'utf8' });
  if (bash.status === 0) return true;

  console.warn('bash build-wasm.sh failed, trying go directly…');
  console.warn(bash.stderr || bash.stdout || bash.error);

  const goRoot = spawnSync('go', ['env', 'GOROOT'], { encoding: 'utf8' });
  if (goRoot.status !== 0) {
    throw new Error('Go is not installed or not on PATH. Install Go to build motohud.wasm.');
  }
  const goroot = goRoot.stdout.trim();
  const wasmExecSrc = join(goroot, 'lib', 'wasm', 'wasm_exec.js');
  copyFileSync(wasmExecSrc, join(repoRoot, 'web', 'emulator', 'wasm_exec.js'));
  copyFileSync(wasmExecSrc, outExec);

  const build = spawnSync(
    'go',
    ['build', '-o', join(repoRoot, 'web', 'emulator', 'motohud.wasm'), './cmd/motohud-wasm'],
    {
      cwd: join(repoRoot, 'pi'),
      env: { ...process.env, GOOS: 'js', GOARCH: 'wasm' },
      encoding: 'utf8',
    },
  );
  if (build.status !== 0) {
    throw new Error(build.stderr || build.stdout || 'go build wasm failed');
  }
  return true;
}

runGoWasmBuild();

const srcWasm = join(repoRoot, 'web', 'emulator', 'motohud.wasm');
const srcExec = join(repoRoot, 'web', 'emulator', 'wasm_exec.js');
if (!existsSync(srcWasm)) {
  throw new Error(`missing ${srcWasm}`);
}
copyFileSync(srcWasm, outWasm);
if (existsSync(srcExec)) copyFileSync(srcExec, outExec);

console.log('wrote', outWasm);
