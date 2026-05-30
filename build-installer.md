# Building the Windows installer

The installer is based on **Inno Setup** (script `installer.iss` in the repo root) and produces `dist\SyntaxChecker-Setup.exe`.

Installed content:

- `syntax-checker.exe` and `syntaxchecker-mcp.exe` in the installation folder (default: `%LOCALAPPDATA%\Programs\SyntaxChecker`)
- `for-ia.md` and `for-agent.md` in the `docs\` subfolder
- An uninstall entry in *Apps & features*
- (Optional, a checkable task in setup) adding the installation folder to the user `PATH`

## Prerequisites

1. **Go 1.22+** on the `PATH` (to build the two exes).
2. **Inno Setup 6** installed. Download: <https://jrsoftware.org/isdl.php>.
   - Make sure `iscc.exe` is on the `PATH`, or pass the full path via `ISCC=...`.
   - Typical path: `C:\Program Files (x86)\Inno Setup 6\ISCC.exe`.
3. **make** (e.g. via `choco install make`, or use the manual command below).

## Build — a single command (PowerShell, no make)

From the repo root:

```powershell
.\build-installer.ps1
# optional:
.\build-installer.ps1 -Version 1.2.3
.\build-installer.ps1 -Iscc "C:\Program Files (x86)\Inno Setup 6\ISCC.exe"
```

The script locates `iscc.exe` on the `PATH` or in the standard installation paths, builds the two exes for `windows/amd64`, then compiles the installer. The version comes from `git describe` (fallback `dev`).

## Build via Mage

```powershell
mage installer
```

What it does:

1. Runs `mage windows` → builds `dist\syntax-checker.exe` and `dist\syntaxchecker-mcp.exe` for `windows/amd64`.
2. Runs `iscc /DMyAppVersion=<git-version> installer.iss` → produces `dist\SyntaxChecker-Setup.exe`.

The version comes from `git describe --tags --always` (fallback `dev`).

If `iscc` is not on the `PATH`, point to it with the `ISCC` environment variable:

```powershell
$env:ISCC = "C:\Program Files (x86)\Inno Setup 6\ISCC.exe"; mage installer
```

## Manual build (without Mage)

```powershell
# 1) build the two exes for Windows
cd apps\checker     ; go build -trimpath -ldflags "-s -w" -o ..\..\dist\syntax-checker.exe .
cd ..\mcp-server    ; go build -trimpath -ldflags "-s -w" -o ..\..\dist\syntaxchecker-mcp.exe .
cd ..\..

# 2) compile the installer
& "C:\Program Files (x86)\Inno Setup 6\ISCC.exe" /DMyAppVersion=1.2.3 installer.iss
```

Output: `dist\SyntaxChecker-Setup.exe`.

## Quick test

1. Run `dist\SyntaxChecker-Setup.exe` on a clean Windows machine (or in a VM).
2. Confirm:
   - The files `syntax-checker.exe`, `syntaxchecker-mcp.exe` are in `{app}`.
   - `docs\for-ia.md` and `docs\for-agent.md` are in `{app}\docs`.
   - If you checked *"Add the installation folder to the user PATH"*: open a **new** PowerShell and verify `where.exe syntax-checker` and `where.exe syntaxchecker-mcp`.
3. Uninstall from *Apps & features*: the `{app}` folder is removed and the `PATH` entry is cleaned up.

## Notes

- The installer runs **per-user** by default (`PrivilegesRequired=lowest`), but shows the dialog to choose *All users* if an admin install is needed. In admin mode it writes the system `PATH` instead of the user one.
- The script handles the `PATH` correctly: no duplicates on install, clean removal on uninstall.
- To change the name/publisher/AppId, edit the `#define`s at the top of `installer.iss`. Do **not** change `AppId` after the first release (it breaks in-place upgrades).
