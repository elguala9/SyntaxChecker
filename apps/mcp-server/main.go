// Command syntaxchecker-mcp exposes the syntax checker as an MCP tool. It invokes the
// `syntax-checker` executable as a subprocess and returns its JSON result.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// CheckInput is the argument schema for the check_syntax tool.
type CheckInput struct {
	FilePath string `json:"file_path" jsonschema:"path of the file to check (absolute, or relative to the server's working directory)"`
	Type     string `json:"type,omitempty" jsonschema:"forced type: json, json5, jsonc, xml, html, yaml, toml, ini, csv, tsv, hcl, markdown, env, properties, proto, graphql, dockerfile, compose, jq, go, ts, tsx, js, jsx, lua, shell, starlark, sql:mysql, sql:postgres, sql:ansi, sql:sqlite, sql:mssql, sql:oracle; empty means auto-detect from the extension"`
	Schema   string `json:"schema,omitempty" jsonschema:"path of a JSON Schema (json, yaml) or a DTD (xml) to validate the document against"`
	Strict   bool   `json:"strict,omitempty" jsonschema:"enable stricter checks where supported (e.g. JSON duplicate keys)"`
}

// checkerBin is resolved once at startup.
var checkerBin = resolveCheckerBin()

func main() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "syntaxchecker",
		Version: "0.1.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "check_syntax",
		Description: "Validate the syntax of a file (JSON, JSON5, JSONC, XML, HTML, YAML, TOML, INI, CSV/TSV, HCL, Markdown, .env, Properties, Protobuf, GraphQL, Dockerfile, jq, the programming languages Go, TypeScript/JavaScript incl. JSX/TSX, Shell/Bash, Lua, Starlark, or SQL). Programming-language checks are parse-only (no type-checking). Optionally validate JSON/YAML against a JSON Schema or XML against a DTD by passing the 'schema' path. Returns whether the file is valid and the list of syntax errors with line/column when available.",
	}, checkSyntax)

	// stdin EOF (client disconnect) is a normal shutdown, not a failure.
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Printf("syntaxchecker-mcp stopped: %v", err)
	}
}

func checkSyntax(ctx context.Context, _ *mcp.CallToolRequest, in CheckInput) (*mcp.CallToolResult, result.CheckResult, error) {
	if strings.TrimSpace(in.FilePath) == "" {
		return nil, result.CheckResult{}, errors.New("file_path is required")
	}

	args := []string{"-f", in.FilePath, "-o", "json"}
	if in.Type != "" {
		args = append(args, "-t", in.Type)
	}
	if in.Schema != "" {
		args = append(args, "-s", in.Schema)
	}
	if in.Strict {
		args = append(args, "--strict")
	}

	cmd := exec.CommandContext(ctx, checkerBin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	exitCode := 0
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			exitCode = ee.ExitCode()
		} else {
			return nil, result.CheckResult{}, fmt.Errorf("cannot run checker %q: %w", checkerBin, err)
		}
	}

	// Exit 2 = internal error (file not found, unsupported type): no JSON on stdout.
	if exitCode == 2 {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = "internal error"
		}
		return nil, result.CheckResult{}, fmt.Errorf("checker: %s", msg)
	}

	// Exit 0 (valid) and 1 (invalid) both print a CheckResult as JSON.
	var res result.CheckResult
	if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
		return nil, result.CheckResult{}, fmt.Errorf("cannot parse checker output: %w (raw: %q)", err, stdout.String())
	}
	return nil, res, nil
}

// resolveCheckerBin locates the checker executable: $CHECKER_BIN, then a binary
// next to this server, then the PATH.
func resolveCheckerBin() string {
	if v := os.Getenv("CHECKER_BIN"); v != "" {
		return v
	}
	name := "syntax-checker"
	if runtime.GOOS == "windows" {
		name = "syntax-checker.exe"
	}
	if exe, err := os.Executable(); err == nil {
		cand := filepath.Join(filepath.Dir(exe), name)
		if _, err := os.Stat(cand); err == nil {
			return cand
		}
	}
	return name
}
