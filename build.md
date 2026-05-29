# Local build

Instructions to compile the `syntax-checker` and `syntaxchecker-mcp` binaries into the `dist/` folder.

Builds are driven by [Mage](https://magefile.org), a Go-native task runner (the build
tasks live in `magefiles/magefile.go`). It works the same on Windows, Linux, and macOS.

## Prerequisites

- Go 1.25+ (`go version`)
- Mage:
  ```bash
  go install github.com/magefile/mage@latest
  ```
  Make sure `$(go env GOPATH)/bin` is on your `PATH` so the `mage` command is found.
- Run all commands from the repository root

## Build for the current platform

```bash
mage build       # or just: mage
```

Produces in `dist/`:
- `syntax-checker` (or `syntax-checker.exe` on Windows)
- `syntaxchecker-mcp` (or `syntaxchecker-mcp.exe` on Windows)

To build only one of them:

```bash
mage checker
mage mcp
```

List every available target:

```bash
mage -l
```

## Cross-platform build

### Windows (`.exe`)

```bash
mage windows
```

Generates `dist/syntax-checker.exe` and `dist/syntaxchecker-mcp.exe` (target `windows/amd64`).

### Linux

```bash
mage linux
```

Generates `dist/syntax-checker` and `dist/syntaxchecker-mcp` (target `linux/amd64`).

## Manual build (without Mage)

If you prefer not to install Mage, from PowerShell:

```powershell
$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

$VERSION   = (git describe --tags --always 2>$null); if (-not $VERSION) { $VERSION = "dev" }
$COMMIT    = (git rev-parse --short HEAD 2>$null);   if (-not $COMMIT)  { $COMMIT  = "none" }
$BUILDDATE = (Get-Date -AsUTC -Format "yyyy-MM-ddTHH:mm:ssZ")

$LDFLAGS = "-s -w -X main.version=$VERSION -X main.commit=$COMMIT -X main.buildDate=$BUILDDATE"

New-Item -ItemType Directory -Force dist | Out-Null
go build -trimpath -ldflags "$LDFLAGS" -o dist/syntax-checker.exe ./apps/checker
go build -trimpath -ldflags "-s -w"    -o dist/syntaxchecker-mcp.exe ./apps/mcp-server
```

## Verify

```bash
./dist/syntax-checker --version
./dist/syntaxchecker-mcp --version
```

## Clean

```bash
mage clean
```

Removes the `dist/` folder.
