# Setup di `AGENTS.md` / `CLAUDE.md` nei tuoi progetti

Questo documento spiega come configurare un file di istruzioni per gli agenti AI (Claude Code, OpenCode, ecc.) in un progetto, in modo che usino automaticamente il MCP `syntaxchecker` per validare file JSON / XML / YAML / TOML / INI / CSV / HCL / Markdown / `.env` / Properties / SQL.

L'installer copia questo file insieme a `for-ia.md` nella cartella di installazione, così è sempre disponibile come riferimento.

## Quale file usare

Gli agenti AI leggono, all'avvio della sessione, un file di istruzioni nella root del progetto. La convenzione dipende dal client:

| Client                | File letto automaticamente |
|-----------------------|----------------------------|
| Claude Code           | `CLAUDE.md` (root del repo, anche annidati nelle sottocartelle) |
| OpenCode              | `AGENTS.md` (root del repo) |
| Cursor / altri agenti | `AGENTS.md` (standard de-facto emergente) |

**Consiglio pratico**: crea un unico `AGENTS.md` con le istruzioni e poi, nei progetti dove usi Claude Code, aggiungi un `CLAUDE.md` che si limita a includere il contenuto di `AGENTS.md` (oppure tienili identici). In questo modo qualunque agente trova le stesse regole.

## Cosa scrivere dentro

Il file deve dire all'agente **quando** invocare `check_syntax` e **come** chiamarlo. Sotto trovi un template minimo, copia-incollabile nella root del progetto.

### Template `AGENTS.md` (e/o `CLAUDE.md`)

```markdown
# Istruzioni per agenti AI in questo progetto

## Validazione sintattica dei file

In questo progetto è disponibile il MCP server **`syntaxchecker`**, che espone
il tool `check_syntax` per JSON, XML, YAML, TOML, INI, CSV/TSV, HCL, Markdown,
`.env`, Properties e SQL (MySQL, PostgreSQL, ANSI, SQLite, SQL Server, Oracle).

**Regole obbligatorie:**

1. Dopo aver creato o modificato un file `.json`, `.xml`, `.yml`/`.yaml`,
   `.toml`, `.ini`/`.cfg`, `.csv`/`.tsv`, `.hcl`/`.tf`, `.md`, `.env`,
   `.properties` o `.sql`, invoca SEMPRE `check_syntax` sul file prima di
   considerare il task concluso.
2. Per i file `.sql` passa SEMPRE il parametro `type` con il dialetto corretto
   (`sql:mysql`, `sql:postgres`, `sql:ansi`, `sql:sqlite`, `sql:mssql`,
   `sql:oracle`). L'auto-detect non sceglie dialetti SQL.
3. Per i file JSON di configurazione passa `strict=true` per intercettare
   chiavi duplicate.
4. Usa SEMPRE path assoluti in `file_path`.
5. Se `valid: false`, correggi il file e ri-valida prima di consegnarlo
   all'utente. Non restituire mai un file rotto.

**Esempi di chiamata:**

```
check_syntax(file_path="C:/proj/config.json", strict=true)
check_syntax(file_path="C:/proj/migrations/001_init.sql", type="sql:postgres")
check_syntax(file_path="C:/proj/deploy/k8s.yaml")
```

Per il riferimento completo del tool vedi `for-ia.md` (installato insieme al
MCP server).
```

## Setup passo-passo in un progetto nuovo

1. Verifica che il MCP `syntaxchecker` sia già configurato nel client AI (vedi `for-ia.md`, sezione *MCP client configuration*). Se non lo è, configuralo prima — il file `AGENTS.md` da solo non installa nulla.
2. Vai nella root del progetto.
3. Crea `AGENTS.md` incollando il template qui sopra. Adatta gli esempi ai path/tipi di file effettivamente presenti nel progetto (es. se non hai SQL, togli la regola sul dialetto).
4. Se usi Claude Code, crea anche `CLAUDE.md` con lo stesso contenuto (oppure una riga `Vedi AGENTS.md.` se l'agente segue i rimandi — Claude Code in genere preferisce contenuto diretto).
5. Committa entrambi i file nel repo: in questo modo le regole valgono per chiunque cloni il progetto, non solo per la tua macchina.
6. Riavvia la sessione dell'agente (nuova conversazione) così che rilegga il file.

## Verifica che funzioni

Apri una nuova conversazione con l'agente nella root del progetto e chiedi qualcosa come:

> "Crea un file `test.json` con una configurazione di esempio."

L'agente, dopo aver scritto il file, dovrebbe spontaneamente invocare `check_syntax` su di esso. Se non lo fa, controlla:

- che il file si chiami esattamente `AGENTS.md` o `CLAUDE.md` (case-sensitive su Linux/macOS);
- che sia nella root del repo (o in una cartella che l'agente sta esplorando);
- che il MCP `syntaxchecker` risulti attivo nel client (in Claude Code: `/mcp`; in OpenCode: lista dei tool disponibili).

## Note per chi prepara l'installer

L'installer dovrebbe:

- Copiare `for-agent.md` (questo file) e `for-ia.md` nella cartella di installazione, sezione `docs/`.
- NON creare automaticamente `AGENTS.md` o `CLAUDE.md` nei progetti dell'utente: sono file per-repo, vanno aggiunti manualmente dove servono. L'installer può però offrire un comando o un menu *"Copia template `AGENTS.md` nella cartella corrente"* per comodità.
