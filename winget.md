# Publishing to winget

This document describes how to publish **SyntaxChecker** to the
[Windows Package Manager](https://learn.microsoft.com/windows/package-manager/)
(`winget`) community repository, and how to ship updates afterwards.

Once published, users install with:

```powershell
winget install Parresia.SyntaxChecker
```

---

## How winget works (in 30 seconds)

- A package lives in the community repo
  [`microsoft/winget-pkgs`](https://github.com/microsoft/winget-pkgs) as a set of
  **YAML manifests** under `manifests/p/Parresia/SyntaxChecker/<version>/`.
- Each release/version is a **new folder** with its own manifests; you never edit
  an old version in place.
- The manifests point at a **stable download URL** (our GitHub Release asset) and
  pin its **SHA256**.
- Publishing/updating = opening a **pull request** to `winget-pkgs`. Microsoft's
  bots validate it and, once merged, the package is available worldwide.

**Package identifier:** `Parresia.SyntaxChecker` (`<Publisher>.<PackageName>`).
This is permanent — choose it once and never change it.

---

## Prerequisites

1. A **published GitHub Release** with the Windows asset and its checksum (the
   `Release` workflow already produces these — see [`release.md`](release.md)):
   - `syntaxchecker-v<version>-windows-amd64.zip`
   - `syntaxchecker-v<version>-windows-amd64.zip.sha256`
   > The download URL must be **immutable**. GitHub Release assets are, so we're good.
2. A **GitHub account with a fork** of `microsoft/winget-pkgs` (the tooling creates
   it for you on first run).
3. A **GitHub Personal Access Token (classic)** with the `public_repo` scope (used
   to push the branch and open the PR). Needed only for the `--submit` step.
4. The publishing tool — pick one:
   - **wingetcreate** (Microsoft official): `winget install Microsoft.WingetCreate`
   - **komac** (community, great for CI): `winget install KomacContributors.Komac`

---

## Which installer type?

The release ships a **zip** containing both executables, so the natural winget
package is a **portable (zip)** package: winget unpacks the zip and registers
PATH aliases for the two binaries. No admin rights, clean uninstall.

| Type | Pros | Cons |
|---|---|---|
| **Portable (zip)** ← recommended | Uses the existing release asset as-is; no admin; auto PATH aliases | The nested folder name embeds the version (handled per-version) |
| **Inno Setup (.exe)** | Native installer, Add/Remove entry | Requires also publishing `SyntaxChecker-Setup.exe` in the Release (see [`build-installer.md`](build-installer.md)); not currently uploaded by the `Release` workflow |

The rest of this guide uses the **portable (zip)** approach.

---

## First-time publish

### Option A — wingetcreate (interactive, recommended)

From any Windows machine:

```powershell
wingetcreate new https://github.com/elguala9/SyntaxChecker/releases/download/v0.1.0/syntaxchecker-v0.1.0-windows-amd64.zip
```

`wingetcreate` downloads the asset, computes the SHA256, and prompts for the
metadata. Answer with:

| Prompt | Value |
|---|---|
| PackageIdentifier | `Parresia.SyntaxChecker` |
| PackageVersion | `0.1.0` (no leading `v`) |
| InstallerType | `zip` |
| NestedInstallerType | `portable` |
| Nested file 1 | `syntaxchecker-v0.1.0-windows-amd64\syntax-checker.exe`, alias `syntax-checker` |
| Nested file 2 | `syntaxchecker-v0.1.0-windows-amd64\syntaxchecker-mcp.exe`, alias `syntaxchecker-mcp` |
| Architecture | `x64` |
| Publisher | `Parresia` |
| PackageName | `SyntaxChecker` |
| License | `MIT` |
| ShortDescription | `Multi-format syntax validator CLI and MCP server.` |
| PackageUrl | `https://github.com/elguala9/SyntaxChecker` |

At the end it validates the manifests and asks to **submit the PR**. Provide your
PAT when prompted (or pass `--token <PAT>`).

### Option B — manual manifests

Create three files under `manifests/p/Parresia/SyntaxChecker/0.1.0/`:

**`Parresia.SyntaxChecker.yaml`** (version):

```yaml
PackageIdentifier: Parresia.SyntaxChecker
PackageVersion: 0.1.0
DefaultLocale: en-US
ManifestType: version
ManifestVersion: 1.6.0
```

**`Parresia.SyntaxChecker.installer.yaml`**:

```yaml
PackageIdentifier: Parresia.SyntaxChecker
PackageVersion: 0.1.0
InstallerType: zip
NestedInstallerType: portable
NestedInstallerFiles:
  - RelativeFilePath: syntaxchecker-v0.1.0-windows-amd64\syntax-checker.exe
    PortableCommandAlias: syntax-checker
  - RelativeFilePath: syntaxchecker-v0.1.0-windows-amd64\syntaxchecker-mcp.exe
    PortableCommandAlias: syntaxchecker-mcp
Installers:
  - Architecture: x64
    InstallerUrl: https://github.com/elguala9/SyntaxChecker/releases/download/v0.1.0/syntaxchecker-v0.1.0-windows-amd64.zip
    InstallerSha256: <PASTE-SHA256-HERE>
ManifestType: installer
ManifestVersion: 1.6.0
```

> Get the SHA256 from the published `.sha256` asset, or compute it:
> `(Get-FileHash .\syntaxchecker-v0.1.0-windows-amd64.zip -Algorithm SHA256).Hash`

**`Parresia.SyntaxChecker.locale.en-US.yaml`**:

```yaml
PackageIdentifier: Parresia.SyntaxChecker
PackageVersion: 0.1.0
PackageLocale: en-US
Publisher: Parresia
PackageName: SyntaxChecker
License: MIT
LicenseUrl: https://github.com/elguala9/SyntaxChecker/blob/master/LICENSE
ShortDescription: Multi-format syntax validator CLI and MCP server.
PackageUrl: https://github.com/elguala9/SyntaxChecker
Tags:
  - syntax
  - validator
  - linter
  - cli
  - mcp
ManifestType: defaultLocale
ManifestVersion: 1.6.0
```

Validate, test the install locally, then submit:

```powershell
winget validate --manifest manifests\p\Parresia\SyntaxChecker\0.1.0
winget install  --manifest manifests\p\Parresia\SyntaxChecker\0.1.0   # local test (enable: winget settings → localManifestFiles)
wingetcreate submit --token <PAT> manifests\p\Parresia\SyntaxChecker\0.1.0
```

---

## Shipping an update (new version)

After the `Release` workflow has published a **new** version (e.g. `v0.2.0`),
update the package — one command:

```powershell
wingetcreate update Parresia.SyntaxChecker `
  --version 0.2.0 `
  --urls https://github.com/elguala9/SyntaxChecker/releases/download/v0.2.0/syntaxchecker-v0.2.0-windows-amd64.zip `
  --submit `
  --token <PAT>
```

`wingetcreate update` pulls the current manifests, bumps the version, re-downloads
the asset, recomputes the SHA256, fixes the nested `RelativeFilePath` (the folder
name carries the version), and opens the PR.

> **komac** equivalent:
> ```powershell
> komac update Parresia.SyntaxChecker --version 0.2.0 `
>   --urls https://github.com/elguala9/SyntaxChecker/releases/download/v0.2.0/syntaxchecker-v0.2.0-windows-amd64.zip `
>   --submit
> ```

---

## Optional — automate the winget PR from CI

You can open the winget update PR automatically when a release is published.
Append a job to `.github/workflows/release.yml` (runs after `release`):

```yaml
  winget:
    name: Submit winget update
    needs: release
    runs-on: windows-latest
    steps:
      - name: Update winget manifest
        run: |
          $ver = "${{ github.ref_name }}".TrimStart('v')
          $url = "https://github.com/elguala9/SyntaxChecker/releases/download/${{ github.ref_name }}/syntaxchecker-${{ github.ref_name }}-windows-amd64.zip"
          komac update Parresia.SyntaxChecker --version $ver --urls $url --submit --token $env:WINGET_TOKEN
        env:
          WINGET_TOKEN: ${{ secrets.WINGET_TOKEN }}
```

Notes:
- `WINGET_TOKEN` must be a **PAT with `public_repo`** that can push to *your fork*
  of `winget-pkgs` — the default `GITHUB_TOKEN` cannot open PRs on another repo.
- Install `komac` first in the job (`winget install KomacContributors.Komac` or
  download the release binary), or use `wingetcreate`.
- The package must already exist (do the **first-time publish** manually once);
  CI only handles **updates**.

---

## Checklist

- [ ] GitHub Release published with `...-windows-amd64.zip` and its `.sha256`
- [x] `License` set to `MIT` in the locale manifest (see `LICENSE`)
- [ ] First version submitted manually and **merged** into `winget-pkgs`
- [ ] `winget install Parresia.SyntaxChecker` works on a clean machine
- [ ] (Optional) CI step wired with a `WINGET_TOKEN` secret for future updates
