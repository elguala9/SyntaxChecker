module github.com/parresia/syntaxchecker/apps/checker

go 1.25.0

require (
	github.com/antlr4-go/antlr/v4 v4.13.1
	github.com/bytebase/parser v0.0.0-20260417075056-57b6ef7a2640
	github.com/hashicorp/hcl/v2 v2.24.0
	github.com/itchyny/gojq v0.12.19
	github.com/joho/godotenv v1.5.1
	github.com/magiconair/properties v1.8.10
	github.com/moby/buildkit v0.30.0
	github.com/parresia/syntaxchecker/pkg/result v0.0.0
	github.com/pelletier/go-toml/v2 v2.3.1
	github.com/pingcap/tidb/pkg/parser v0.0.0-20260527114842-beb12a7923d3
	github.com/rqlite/sql v0.0.0-20260224021119-1b2524a41372
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2
	github.com/tidwall/jsonc v0.3.3
	github.com/titanous/json5 v1.0.0
	github.com/vektah/gqlparser/v2 v2.5.33
	github.com/wasilibs/go-pgquery v0.0.0-20260526011917-40df1ddb6e56
	github.com/yoheimuta/go-protoparser/v4 v4.14.2
	github.com/yuin/goldmark v1.8.2
	golang.org/x/net v0.55.0
	gopkg.in/ini.v1 v1.67.2
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/containerd/typeurl/v2 v2.2.3 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/itchyny/timefmt-go v0.1.8 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/pganalyze/pg_query_go/v6 v6.2.2 // indirect
	github.com/pingcap/errors v0.11.5-0.20250523034308-74f78ae071ee // indirect
	github.com/pingcap/failpoint v0.0.0-20240528011301-b51a646c7c86 // indirect
	github.com/pingcap/log v1.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/tetratelabs/wazero v1.11.0 // indirect
	github.com/tonistiigi/go-csvvalue v0.0.0-20240814133006-030d3b2625d0 // indirect
	github.com/wasilibs/wazero-helpers v0.0.0-20250123031827-cd30c44769bb // indirect
	github.com/zclconf/go-cty v1.16.3 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/exp v0.0.0-20250911091902-df9299821621 // indirect
	golang.org/x/mod v0.35.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	golang.org/x/tools v0.44.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)

replace github.com/parresia/syntaxchecker/pkg/result => ../../pkg/result
