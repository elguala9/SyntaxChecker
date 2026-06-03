# winget publish — status & known issues

Working notes for publishing **Parresia.SyntaxChecker** to winget.

## Where things stand

- **`0.1.1` is LIVE** on winget (`winget show Parresia.SyntaxChecker` returns it).
  The first-time PR to `microsoft/winget-pkgs` was submitted and merged.
- **`0.2.0`** GitHub release is published with the Windows asset and checksum:
  - zip: `syntaxchecker-v0.2.0-windows-amd64.zip`
  - SHA256: `0A7FBE31009E2D584BB7381E8AC10190040F9A0F86533DA6B6BD4FB2C8C661D6` (verified)
  - The `0.2.0` release includes the Docker/Podman Compose checker (absent from `0.1.1`).
- Reference manifests for `0.2.0` live under `winget-manifests/0.2.0/` and pass
  `winget validate`. **The winget-pkgs update PR for `0.2.0` is not submitted yet.**

## Shipping the 0.2.0 update — commands

Because the package already exists on winget, use the non-interactive `update`
flow (it does **not** hit the `new`-flow parse problem below). Needs a
**PAT classic** with the `public_repo` scope.

1. Dry run (no submit) to inspect the generated manifests:
   ```powershell
   wingetcreate update Parresia.SyntaxChecker `
     --version 0.2.0 `
     --urls "https://github.com/elguala9/SyntaxChecker/releases/download/v0.2.0/syntaxchecker-v0.2.0-windows-amd64.zip"
   ```
   Confirm the nested `RelativeFilePath` entries were bumped to
   `syntaxchecker-v0.2.0-windows-amd64\...` (the folder name carries the version).

2. Submit the PR:
   ```powershell
   wingetcreate update Parresia.SyntaxChecker `
     --version 0.2.0 `
     --urls "https://github.com/elguala9/SyntaxChecker/releases/download/v0.2.0/syntaxchecker-v0.2.0-windows-amd64.zip" `
     --submit `
     --token <PAT>
   ```

Alternatively, submit the validated reference manifests in this repo directly:
```powershell
wingetcreate submit --token <PAT> "winget-manifests\0.2.0"
```

After the Microsoft bots validate and merge, users get it with:
```powershell
winget upgrade Parresia.SyntaxChecker
```

## Known issue — `wingetcreate new` fails on portable zips

`wingetcreate new <zip-url>` errors with "Could not parse the package" because it
tries to analyze the selected `.exe` files as installers; our binaries are plain
Go executables (`CGO_ENABLED=0`, no installer metadata). This affects only the
interactive `new` flow. Both `wingetcreate update` and `wingetcreate submit`
(used above) avoid it, since they don't re-analyze the binaries.

## Notes / loose ends

- `winget validate` checks manifest **schema** only, not the SHA256 vs download.
  The bots verify the hash in the PR (and we verified it locally — MATCH).
- Keeping `winget-manifests/` in the repo is handy as a reference for future
  `wingetcreate update` runs.
