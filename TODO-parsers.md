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

- [x] **JSON5 / JSONC** — `.json5`, `.jsonc` — lib: `titanous/json5` (JSON5), `tidwall/jsonc` (JSONC→JSON, poi riuso `JSON{}`)
- [x] **HTML** — `.html`, `.htm` — lib: `golang.org/x/net/html`. Lenient = parse tollerante HTML5 (di fatto sempre valido, come Markdown); `--strict` = well-formedness in stile XHTML (tag bilanciati/annidati, void element gestiti). Vedi `html.go`.
- [x] **GraphQL** — `.graphql`, `.gql` — lib: `vektah/gqlparser/v2` (valido se parsa come schema SDL o come query)
- [x] **Protobuf** — `.proto` — lib: `github.com/yoheimuta/go-protoparser/v4` (modalità non-permissive)
- [x] **Dockerfile** — `Dockerfile`, `Containerfile`, `*.dockerfile` — lib: `moby/buildkit` frontend parser (parse strutturale + `instructions.ParseInstruction` per istruzioni sconosciute/argomenti malformati). Rilevamento per nome file in `detectType`. Vedi `dockerfile.go`.
- [x] **JQ** — `.jq` — lib: `github.com/itchyny/gojq` (parse-only delle espressioni). Vedi `jq.go`.
  - **JSONPath**: fuori scope — nessuna libreria Go matura con semantica standardizzata univoca; le espressioni non sono un formato-file tipico per questo tool.

## Estensioni dei formati esistenti
- [x] **XML + DTD** — `SchemaValidator` su `XML{}` (`xml_dtd.go`): valida elementi dichiarati, content model (membership, EMPTY, element-content vs #PCDATA), attributi `#REQUIRED`/`#FIXED`/enumerazioni, attributi non dichiarati. Limiti documentati: niente ordine/cardinalità del content model, niente espansione entità, niente ID/IDREF.
  - **XSD / RelaxNG**: NON supportati di proposito. Non esiste un validatore puro-Go maturo; l'unica opzione production-grade (cgo + libxml2) romperebbe il binario statico self-contained e l'installer Windows (stessa motivazione documentata in `xml.go`).
- [x] **YAML strict** — `--strict` rifiuta i tag custom (es. `!Point`, `!!python/object`) che non sono rappresentabili come dati YAML/JSON portabili; anchor/alias e tag standard restano validi. Vedi `yaml.go` (`yamlCustomTags`).
