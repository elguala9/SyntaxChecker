# winget publish — status & known issues (v0.1.1)

Working notes for publishing **Parresia.SyntaxChecker** to winget. Pick up from
the **Resume here** section.

## Where things stand

- Release **`v0.1.1`** is published on GitHub and its zip now bundles the docs:
  ```
  syntaxchecker-v0.1.1-windows-amd64/
  ├── for-agent.md
  ├── for-ia.md
  ├── README.md
  ├── syntax-checker.exe
  └── syntaxchecker-mcp.exe
  ```
  SHA256 of the zip: `B0FAABC2C8A3E7E7E1D9BEADAFDE136C5B24A3D52B08BA389CE7509B9792BB09`
- The `release.yml` workflow was updated to copy `README.md for-ia.md for-agent.md`
  into the archive (only future releases; `v0.1.0` does **not** have the docs).
- `for-ia.md` got a new section "Installed via winget (portable)" documenting the
  PATH-alias setup (`claude mcp add syntaxchecker syntaxchecker-mcp`).
- Manual winget manifests were generated and **validated OK** under
  `winget-manifests/0.1.1/` (3 files). Submission to winget-pkgs is **not done yet**.

## Known issue — `wingetcreate new` fails on portable zips

Command tried:
```powershell
wingetcreate new https://github.com/elguala9/SyntaxChecker/releases/download/v0.1.1/syntaxchecker-v0.1.1-windows-amd64.zip
```
After selecting the two `.exe` files it errors with:
> Non è stato possibile analizzare il pacchetto da [...zip]
> (Could not parse the package)

**Cause:** `wingetcreate new` tries to analyze the selected files as *installers*
to infer metadata. Our binaries are plain Go executables (built with
`CGO_ENABLED=0`, no installer metadata), so the analysis step fails. This is a
limitation of the interactive `new` flow with **zip + portable** packages, not a
version bug — reproduced on WingetCreate **1.12.8** (latest).

**Workaround (the one we adopted):** skip `wingetcreate new`, write the manifests
by hand, then submit them with `wingetcreate submit` (which does **not** re-analyze
the binaries). The validated manifests already live in `winget-manifests/0.1.1/`.

Also tried first: launched `wingetcreate new` with the wrong URL (`v0.1.0`) by
mistake — that zip lacks the docs. Always use the **`v0.1.1`** URL.

## Resume here — submit the PR

1. (Optional) local install test — run once from an **admin** shell:
   ```powershell
   winget settings --enable LocalManifestFiles
   winget install --manifest "C:\Users\lgualandi\Documents\Development\Parresia\SyntaxChecker\winget-manifests\0.1.1"
   ```
   Then in a **new** shell: `syntax-checker --help` and `syntaxchecker-mcp --help`.

2. Submit the PR to `microsoft/winget-pkgs` (needs a **PAT classic** with the
   `public_repo` scope):
   ```powershell
   wingetcreate submit --token <PAT> "C:\Users\lgualandi\Documents\Development\Parresia\SyntaxChecker\winget-manifests\0.1.1"
   ```
   This forks/updates `microsoft/winget-pkgs`, pushes a branch, and opens the PR.
   Microsoft bots validate (including the real SHA256 vs the download). Once merged:
   ```powershell
   winget install Parresia.SyntaxChecker
   ```

## Notes / loose ends

- `winget validate` only checks the manifest **schema**, not that the SHA256
  matches the download. The bots verify the hash in the PR.
- `winget.md` still shows `0.1.0` in its examples — update to `0.1.1` for
  consistency (cosmetic).
- Decide whether to commit `winget-manifests/` to the repo (handy reference for
  future `wingetcreate update` runs) or keep it untracked.
- Future updates (after a new release) should work with the non-interactive
  `wingetcreate update` / `komac update` commands in `winget.md` — those don't hit
  the `new`-flow parse problem.
