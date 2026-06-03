#!/usr/bin/env node
// Launcher for the syntaxchecker MCP server.
//
// Claude Code always ships a Node runtime, so this single cross-platform script
// replaces per-OS shell launchers. On first use it downloads the platform's
// release archive from GitHub, verifies its SHA-256, extracts the binaries into a
// per-version cache, and then execs `syntaxchecker-mcp` with stdio inherited so
// the MCP stdio transport flows straight through to Claude Code.

import { spawn, execFileSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import {
  existsSync, mkdirSync, mkdtempSync, readFileSync, readdirSync,
  renameSync, rmSync, statSync, writeFileSync, chmodSync,
} from 'node:fs';
import { homedir, tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const REPO = 'elguala9/SyntaxChecker';
const HERE = dirname(fileURLToPath(import.meta.url));

function log(msg) {
  // MCP uses stdout for the protocol; all diagnostics must go to stderr.
  process.stderr.write(`[syntaxchecker] ${msg}\n`);
}

function fail(msg) {
  log(msg);
  process.exit(1);
}

// --- Resolve the target release ------------------------------------------------

const pluginManifest = join(HERE, '..', '.claude-plugin', 'plugin.json');
const version = JSON.parse(readFileSync(pluginManifest, 'utf8')).version;
const tag = `v${version}`;

const OS_MAP = { win32: 'windows', linux: 'linux', darwin: 'darwin' };
const ARCH_MAP = { x64: 'amd64', arm64: 'arm64' };
const os = OS_MAP[process.platform];
const arch = ARCH_MAP[process.arch];

// The release workflow currently builds only these targets.
const SUPPORTED = new Set(['windows-amd64', 'linux-amd64']);
if (!os || !arch || !SUPPORTED.has(`${os}-${arch}`)) {
  if (process.platform === 'darwin') {
    fail(
      'macOS is not supported: no prebuilt binary is published for Darwin. ' +
      `Build from source (https://github.com/${REPO}) and set CHECKER_BIN.`,
    );
  }
  fail(
    `no prebuilt binary for ${process.platform}/${process.arch}. ` +
    `Supported: ${[...SUPPORTED].join(', ')}. ` +
    `Build from source (https://github.com/${REPO}) and set CHECKER_BIN.`,
  );
}

const isWin = os === 'windows';
const exe = (name) => (isWin ? `${name}.exe` : name);
const assetBase = `syntaxchecker-${tag}-${os}-${arch}`;
const assetName = `${assetBase}.${isWin ? 'zip' : 'tar.gz'}`;

// --- Cache layout (per version, so a version bump triggers a fresh fetch) ------

const cacheHome = isWin
  ? (process.env.LOCALAPPDATA || join(homedir(), 'AppData', 'Local'))
  : (process.env.XDG_CACHE_HOME || join(homedir(), '.cache'));
const cacheDir = join(cacheHome, 'syntaxchecker', tag);
const mcpBin = join(cacheDir, exe('syntaxchecker-mcp'));
const checkerBin = join(cacheDir, exe('syntax-checker'));

// --- Download + verify + extract (only if not already cached) ------------------

async function ensureBinaries() {
  if (existsSync(mcpBin) && existsSync(checkerBin)) return;

  const base = `https://github.com/${REPO}/releases/download/${tag}`;
  log(`fetching ${assetName} (first run for ${tag})...`);

  const archiveBuf = await download(`${base}/${assetName}`);
  await verifyChecksum(archiveBuf, `${base}/${assetName}.sha256`);

  const work = mkdtempSync(join(tmpdir(), 'syntaxchecker-'));
  try {
    const archivePath = join(work, assetName);
    writeFileSync(archivePath, archiveBuf);

    const outDir = join(work, 'out');
    mkdirSync(outDir);
    extract(archivePath, outDir);

    const found = {
      mcp: findFile(outDir, exe('syntaxchecker-mcp')),
      checker: findFile(outDir, exe('syntax-checker')),
    };
    if (!found.mcp || !found.checker) {
      fail(`archive ${assetName} did not contain the expected binaries`);
    }

    mkdirSync(cacheDir, { recursive: true });
    installBinary(found.mcp, mcpBin);
    installBinary(found.checker, checkerBin);
    log(`installed binaries to ${cacheDir}`);
  } finally {
    rmSync(work, { recursive: true, force: true });
  }
}

async function download(url) {
  const res = await fetch(url, { redirect: 'follow' });
  if (!res.ok) fail(`download failed (${res.status}) for ${url}`);
  return Buffer.from(await res.arrayBuffer());
}

async function verifyChecksum(buf, url) {
  const res = await fetch(url, { redirect: 'follow' });
  if (!res.ok) {
    log(`warning: checksum file unavailable (${res.status}), skipping verification`);
    return;
  }
  // sha256sum/certutil output: "<hex>  <filename>"; take the first hex token.
  const expected = (await res.text()).trim().split(/\s+/)[0].toLowerCase();
  const actual = createHash('sha256').update(buf).digest('hex');
  if (expected && expected !== actual) {
    fail(`checksum mismatch: expected ${expected}, got ${actual}`);
  }
}

function extract(archivePath, outDir) {
  // bsdtar (Windows 10+/macOS) reads both zip and tar.gz; GNU tar (Linux) reads
  // the tar.gz we ship for Linux. Fall back to PowerShell for zip if tar fails.
  try {
    const args = assetName.endsWith('.zip')
      ? ['-xf', archivePath, '-C', outDir]
      : ['-xzf', archivePath, '-C', outDir];
    execFileSync('tar', args, { stdio: 'ignore' });
  } catch (err) {
    if (isWin && assetName.endsWith('.zip')) {
      execFileSync('powershell', [
        '-NoProfile', '-Command',
        `Expand-Archive -Path '${archivePath}' -DestinationPath '${outDir}' -Force`,
      ], { stdio: 'ignore' });
    } else {
      throw err;
    }
  }
}

function findFile(root, name) {
  for (const entry of readdirSync(root, { withFileTypes: true })) {
    const full = join(root, entry.name);
    if (entry.isDirectory()) {
      const hit = findFile(full, name);
      if (hit) return hit;
    } else if (entry.name === name) {
      return full;
    }
  }
  return null;
}

function installBinary(src, dest) {
  // Move into place, then mark executable on POSIX.
  try {
    renameSync(src, dest);
  } catch {
    // Cross-device rename can fail; copy instead.
    writeFileSync(dest, readFileSync(src));
  }
  if (!isWin) chmodSync(dest, 0o755);
}

// --- Run -----------------------------------------------------------------------

await ensureBinaries();

const child = spawn(mcpBin, [], {
  stdio: 'inherit',
  env: { ...process.env, CHECKER_BIN: checkerBin },
});
child.on('error', (err) => fail(`failed to start syntaxchecker-mcp: ${err.message}`));
child.on('exit', (code, signal) => {
  if (signal) process.kill(process.pid, signal);
  else process.exit(code ?? 0);
});
