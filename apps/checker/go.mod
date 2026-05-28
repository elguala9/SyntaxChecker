module github.com/parresia/syntaxchecker/apps/checker

go 1.25.0

require (
	github.com/antlr4-go/antlr/v4 v4.13.1
	github.com/bytebase/parser v0.0.0-20260417075056-57b6ef7a2640
	github.com/parresia/syntaxchecker/pkg/result v0.0.0
	github.com/pingcap/tidb/pkg/parser v0.0.0-20260527114842-beb12a7923d3
	github.com/rqlite/sql v0.0.0-20260224021119-1b2524a41372
	github.com/wasilibs/go-pgquery v0.0.0-20260526011917-40df1ddb6e56
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/pganalyze/pg_query_go/v6 v6.2.2 // indirect
	github.com/pingcap/errors v0.11.5-0.20250523034308-74f78ae071ee // indirect
	github.com/pingcap/failpoint v0.0.0-20240528011301-b51a646c7c86 // indirect
	github.com/pingcap/log v1.1.0 // indirect
	github.com/tetratelabs/wazero v1.11.0 // indirect
	github.com/wasilibs/wazero-helpers v0.0.0-20250123031827-cd30c44769bb // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)

replace github.com/parresia/syntaxchecker/pkg/result => ../../pkg/result
