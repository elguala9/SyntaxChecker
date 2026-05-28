# Creating a new release

Release publishing is automated by the `.github/workflows/release.yml` workflow, which builds the binaries for **Linux** and **Windows** and creates a GitHub Release with `.tar.gz` / `.zip` archives and their `.sha256` checksums.

There are two ways to trigger it.

---

## Method 1 — Git tag (recommended)

The workflow runs automatically when you push a tag starting with `v`.

### 1. Make sure `master` is clean and up to date

```bash
git checkout master
git pull --ff-only
git status        # must be clean
```

### 2. Create the tag

Follow [SemVer](https://semver.org/): `vMAJOR.MINOR.PATCH` (e.g. `v1.2.0`).

```bash
git tag -a v1.2.0 -m "Release v1.2.0"
```

To list existing tags:

```bash
git tag --sort=-v:refname | head
```

### 3. Push the tag

```bash
git push origin v1.2.0
```

> When the tag is pushed the `Release` workflow starts automatically. Track its progress on GitHub → **Actions** → run "Release".

### 4. Verify the release

Once the workflow finishes, the release is available on GitHub → **Releases** with the following assets:

- `syntaxchecker-v1.2.0-linux-amd64.tar.gz` (+ `.sha256`)
- `syntaxchecker-v1.2.0-windows-amd64.zip` (+ `.sha256`)

Release notes are auto-generated from PRs/commits since the previous release.

---

## Method 2 — Manual trigger (workflow_dispatch)

Useful if you want to re-run the build for an existing tag or create a release without creating the tag from the CLI.

### Via UI

1. GitHub → **Actions** → workflow **Release** → **Run workflow**
2. Enter the `tag` (e.g. `v1.2.0`)
3. **Run workflow**

### Via `gh` CLI

```bash
gh workflow run release.yml -f tag=v1.2.0
```

Monitor:

```bash
gh run watch
gh run list --workflow=release.yml
```

---

## Cancelling / redoing a release

If something goes wrong and you need to redo the tag:

```bash
# delete the tag locally and on the remote
git tag -d v1.2.0
git push origin :refs/tags/v1.2.0

# delete the GitHub release (if it was created)
gh release delete v1.2.0 --yes

# recreate the tag as in Method 1
git tag -a v1.2.0 -m "Release v1.2.0"
git push origin v1.2.0
```

> Avoid re-tagging versions already published to consumers: prefer releasing a new patch (`v1.2.1`).

---

## Pre-release checklist

- [ ] `master` up to date, CI green
- [ ] `make test` and `make lint` pass locally
- [ ] `make build-windows` and `make build-linux` work (see `build.md`)
- [ ] Version chosen following SemVer
- [ ] Annotated tag (`git tag -a`) with a message
- [ ] Tag pushed → `Release` workflow green on GitHub Actions
- [ ] Release visible on GitHub with all 4 assets (2 archives + 2 `.sha256`)
