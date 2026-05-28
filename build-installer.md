# Build dell'installer Windows

L'installer è basato su **Inno Setup** (script `installer.iss` nella root del repo) e produce `dist\SyntaxChecker-Setup.exe`.

Contenuto installato:

- `syntax-checker.exe` e `syntaxchecker-mcp.exe` nella cartella di installazione (default: `%LOCALAPPDATA%\Programs\SyntaxChecker`)
- `for-ia.md` e `for-agent.md` nella sottocartella `docs\`
- Voce di disinstallazione in *App e funzionalità*
- (Opzionale, task spuntabile in setup) aggiunta della cartella di installazione al `PATH` utente

## Prerequisiti

1. **Go 1.22+** nel `PATH` (per buildare i due exe).
2. **Inno Setup 6** installato. Download: <https://jrsoftware.org/isdl.php>.
   - Verifica che `iscc.exe` sia nel `PATH`, oppure passa il path completo via `ISCC=...`.
   - Tipico path: `C:\Program Files (x86)\Inno Setup 6\ISCC.exe`.
3. **make** (es. via `choco install make`, o usa il comando manuale qui sotto).

## Build — un solo comando (PowerShell, senza make)

Dalla root del repo:

```powershell
.\build-installer.ps1
# opzionale:
.\build-installer.ps1 -Version 1.2.3
.\build-installer.ps1 -Iscc "C:\Program Files (x86)\Inno Setup 6\ISCC.exe"
```

Lo script trova `iscc.exe` nel `PATH` o nei path standard di installazione, builda i due exe per `windows/amd64`, poi compila l'installer. Versione presa da `git describe` (fallback `dev`).

## Build via make (se hai make installato)

```powershell
make installer
```

Cosa fa:

1. Lancia `make build-windows` → builda `dist\syntax-checker.exe` e `dist\syntaxchecker-mcp.exe` per `windows/amd64`.
2. Lancia `iscc /DMyAppVersion=<git-version> installer.iss` → produce `dist\SyntaxChecker-Setup.exe`.

La versione viene presa da `git describe --tags --always` (fallback `dev`). Per forzarla:

```powershell
make installer VERSION=1.2.3
```

Se `iscc` non è nel `PATH`:

```powershell
make installer ISCC="C:\Program Files (x86)\Inno Setup 6\ISCC.exe"
```

## Build manuale (senza make)

```powershell
# 1) build dei due exe per Windows
cd apps\checker     ; go build -trimpath -ldflags "-s -w" -o ..\..\dist\syntax-checker.exe .
cd ..\mcp-server    ; go build -trimpath -ldflags "-s -w" -o ..\..\dist\syntaxchecker-mcp.exe .
cd ..\..

# 2) compila l'installer
& "C:\Program Files (x86)\Inno Setup 6\ISCC.exe" /DMyAppVersion=1.2.3 installer.iss
```

Output: `dist\SyntaxChecker-Setup.exe`.

## Test rapido

1. Esegui `dist\SyntaxChecker-Setup.exe` su una macchina Windows pulita (o in una VM).
2. Conferma:
   - I file `syntax-checker.exe`, `syntaxchecker-mcp.exe` sono in `{app}`.
   - `docs\for-ia.md` e `docs\for-agent.md` sono in `{app}\docs`.
   - Se hai spuntato *"Aggiungi la cartella di installazione al PATH utente"*: apri una **nuova** PowerShell e verifica `where.exe syntax-checker` e `where.exe syntaxchecker-mcp`.
3. Disinstalla da *App e funzionalità*: la cartella `{app}` viene rimossa e l'entry nel `PATH` utente viene ripulita.

## Note

- L'installer gira di default **per-utente** (`PrivilegesRequired=lowest`), ma mostra il dialog per scegliere *Tutti gli utenti* se serve admin install. In modalità admin scrive il `PATH` di sistema invece di quello utente.
- Lo script gestisce correttamente il `PATH`: niente duplicati in install, rimozione pulita in uninstall.
- Per cambiare nome/produttore/AppId modifica i `#define` in cima a `installer.iss`. **Non** cambiare `AppId` dopo il primo rilascio (rompe l'upgrade in-place).
