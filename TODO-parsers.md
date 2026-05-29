# TODO — Parser da aggiungere

Elenco dei formati da supportare. Per **ogni** parser servono 4 interventi:

1. Nuovo file `apps/checker/checkers/<nome>.go` con uno struct che implementa `Validator`
   (`Check(data []byte, strict bool) []result.SyntaxError`) — opzionalmente `SchemaValidator`.
2. Registrazione nello switch `validatorFor` (`apps/checker/main.go`).
3. Mappatura estensione in `detectType` (`apps/checker/main.go`).
4. Test (`<nome>_test.go`) + sample in `test-samples/<nome>/` (validi e `*_not_correct.*`).

Pattern di riferimento: `apps/checker/checkers/yaml.go` e `json.go`.

---

## Già supportati
- [x] JSON (+ validazione JSON Schema)
- [x] XML
- [x] YAML (+ validazione JSON Schema)
- [x] SQL — MySQL, PostgreSQL, SQLite, MSSQL/T-SQL, Oracle/PL-SQL, ANSI

---

## Facili / consigliati
Parser maturi o stdlib, mapping errori semplice.

- [x] **TOML** — `.toml` — lib: `pelletier/go-toml/v2` (riga+colonna, duplicati)
- [x] **INI** — `.ini`, `.cfg` — lib: `gopkg.in/ini.v1`
- [x] **CSV/TSV** — `.csv`, `.tsv` — lib: `encoding/csv` (stdlib); `--strict` = conteggio campi costante
- [x] **HCL** — `.hcl`, `.tf` — lib: `github.com/hashicorp/hcl/v2` (riga+colonna)
- [x] **Markdown** — `.md`, `.markdown` — lib: `goldmark` (parse-only: Markdown non ha sintassi invalida)
- [x] **.env** — `.env` — lib: `github.com/joho/godotenv`
- [x] **Properties** — `.properties` — lib: `magiconair/properties` (escape `\uXXXX`, riferimenti circolari)

## Medi
Più lavoro su mapping errori / posizione (riga/colonna).

- [ ] **JSON5 / JSONC** — `.json5`, `.jsonc` — JSON con commenti/trailing comma
- [ ] **HTML** — `.html` — lib: `golang.org/x/net/html` (tollerante) o validatore strict
- [ ] **GraphQL** — `.graphql`, `.gql` — lib: `vektah/gqlparser`
- [x] **Protobuf** — `.proto` — lib: `github.com/yoheimuta/go-protoparser/v4` (modalità non-permissive)
- [ ] **Dockerfile** — `Dockerfile` — lib: `moby/buildkit` frontend parser
- [ ] **JSONPath / JQ** — espressioni — validazione sintassi query

## Estensioni dei formati esistenti
- [ ] **XML + XSD / DTD / RelaxNG** — implementare `SchemaValidator` su `XML{}` (interfaccia già presente in `checker.go`)
- [ ] **YAML strict** — tag custom / anchors in strict mode (JSON Schema già fatto)
