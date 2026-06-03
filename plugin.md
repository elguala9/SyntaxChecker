# SyntaxChecker — Claude Code plugin

This repository doubles as a Claude Code **plugin marketplace**. The `syntaxchecker`
plugin adds the `check_syntax` MCP tool. The platform binaries are downloaded from
GitHub Releases on first use (no winget/dnf/apt required).

- Marketplace manifest: `.claude-plugin/marketplace.json`
- Plugin: `plugin/`
  - `plugin/.claude-plugin/plugin.json` — manifest (the `version` field drives updates)
  - `plugin/.mcp.json` — runs `node bin/launch.mjs`
  - `plugin/bin/launch.mjs` — downloads/verifies/extracts the release, then runs `syntaxchecker-mcp`

## Platform support

| Platform | Status |
| --- | --- |
| Windows x64 | supported |
| Linux x64 | supported |
| macOS | not supported (no prebuilt binary) |
| other (arm64, etc.) | not supported |

On unsupported platforms the launcher exits with a clear message.

---

## 1. Publish (maintainer)

There is **no dedicated "publish" command** (no `npm publish` equivalent). A Claude Code
plugin is published simply by pushing the files to a Git repo that users can reach. The
marketplace is this repository itself.

### First publication

Make sure the plugin files live on the repository's **default branch** (so users can add
the marketplace as `elguala9/SyntaxChecker` without a `@branch` suffix):

```powershell
git checkout main
git pull origin main
git merge develop          # bring in the plugin files
git push origin main       # <-- this push IS the publication
git checkout develop       # (optional) go back to working branch
```

That's it. The plugin is now installable by anyone.

### Releasing a new version (the order matters)

The launcher downloads the release whose tag matches `plugin.json`'s `version`
(`version: 0.1.1` -> release tag `v0.1.1`). Because the `version` field is set explicitly,
**pushing new commits alone does nothing** until you bump it — and the matching release
must exist *before* users update, or the download will fail.

Do it in this order:

1. **Cut the release first** so the binaries exist:
   ```powershell
   git tag v0.1.2
   git push origin v0.1.2     # triggers .github/workflows/release.yml -> publishes assets
   ```
2. **Bump the plugin version** to the same number in `plugin/.claude-plugin/plugin.json`:
   ```json
   { "version": "0.1.2" }
   ```
3. **Push the version bump** to the default branch:
   ```powershell
   git add plugin/.claude-plugin/plugin.json
   git commit -m "chore(plugin): bump to 0.1.2"
   git push origin main
   ```

Now `/plugin marketplace update` will surface `0.1.2` to users, and the launcher will fetch
the matching binaries automatically.

> Tip: keep the git tag and `plugin.json` `version` in sync. If `plugin.json` points to a
> tag that has no published release, the first `check_syntax` call fails to download.

---

## 2. Install (user)

Inside Claude Code:

```text
/plugin marketplace add elguala9/SyntaxChecker
/plugin install syntaxchecker@syntaxchecker
```

- The first command registers the marketplace (clones the repo's default branch).
- The second installs the plugin. `@syntaxchecker` is the marketplace name; you can also
  use the interactive `/plugin` menu (Discover tab) instead.

On the first `check_syntax` call the launcher downloads
`syntaxchecker-<version>-<os>-<arch>` from GitHub Releases, verifies its SHA-256, and
caches the binaries:

- Windows: `%LOCALAPPDATA%\syntaxchecker\v<version>\`
- Linux: `${XDG_CACHE_HOME:-$HOME/.cache}/syntaxchecker/v<version>/`

Subsequent runs reuse the cache (no network needed).

### Pin to a branch (optional)

To track a non-default branch:

```text
/plugin marketplace add elguala9/SyntaxChecker@develop
```

---

## 3. Update (user)

There is **no `/plugin update <name>`** command. Updates come from refreshing the
marketplace, which git-pulls this repo:

```text
/plugin marketplace update syntaxchecker
/reload-plugins
```

- `/plugin marketplace update syntaxchecker` pulls the latest marketplace/plugin
  definitions. A new version appears only if the maintainer bumped `plugin.json`'s
  `version` (see Publish above).
- `/reload-plugins` applies the change in the current session.
- Because the binary cache is keyed by version, moving to a new version triggers a fresh,
  automatic download of the matching release — nothing to clean up by hand.

Auto-update can be toggled per-marketplace in the **Marketplaces** tab of the `/plugin`
menu.

---

## 4. Manage (user)

```text
/plugin disable syntaxchecker@syntaxchecker     # keep installed, turn off
/plugin enable  syntaxchecker@syntaxchecker     # turn back on
/plugin uninstall syntaxchecker@syntaxchecker   # remove
```

Or use the **Installed** tab of the interactive `/plugin` menu.

---

## Quick reference

| Goal | Where | Command |
| --- | --- | --- |
| Publish / update the plugin | terminal | `git push origin main` (after bumping version + release) |
| Add the marketplace | Claude Code | `/plugin marketplace add elguala9/SyntaxChecker` |
| Install | Claude Code | `/plugin install syntaxchecker@syntaxchecker` |
| Update | Claude Code | `/plugin marketplace update syntaxchecker` then `/reload-plugins` |
| Uninstall | Claude Code | `/plugin uninstall syntaxchecker@syntaxchecker` |
