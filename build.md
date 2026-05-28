# Local build

Instructions to compile the `checker` and `mcp-server` binaries into the `dist/` folder.

## Prerequisites

- Go 1.25+ (`go version`)
- `make` available in your shell
  - Windows: use Git Bash, WSL, or install `make` (e.g. `choco install make`)
- Run all commands from the repository root

## Build for the current platform

```bash
make build
```

Produces in `dist/`:
- `checker` (or `checker.exe` on Windows)
- `mcp-server` (or `mcp-server.exe` on Windows)

To build only one of them:

```bash
make build-checker
make build-mcp
```

## Cross-platform build

### Windows (`.exe`)

```bash
make build-windows
```

Generates `dist/checker.exe` and `dist/mcp-server.exe` (target `windows/amd64`).

### Linux

```bash
make build-linux
```

Generates `dist/checker` and `dist/mcp-server` (target `linux/amd64`).

## Manual build (without make)

If `make` is not available, from PowerShell:

```powershell
$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"

$VERSION   = (git describe --tags --always 2>$null); if (-not $VERSION) { $VERSION = "dev" }
$COMMIT    = (git rev-parse --short HEAD 2>$null);   if (-not $COMMIT)  { $COMMIT  = "none" }
$BUILDDATE = (Get-Date -AsUTC -Format "yyyy-MM-ddTHH:mm:ssZ")

$LDFLAGS = "-s -w -X main.version=$VERSION -X main.commit=$COMMIT -X main.buildDate=$BUILDDATE"

New-Item -ItemType Directory -Force dist | Out-Null
Push-Location apps/checker
go build -trimpath -ldflags "$LDFLAGS" -o ../../dist/checker.exe .
Pop-Location
Push-Location apps/mcp-server
go build -trimpath -ldflags "-s -w" -o ../../dist/mcp-server.exe .
Pop-Location
```

## Verify

```bash
./dist/checker --version
./dist/mcp-server --version
```

## Clean

```bash
make clean
```

Removes the `dist/` folder.
