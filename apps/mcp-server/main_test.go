package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/parresia/syntaxchecker/pkg/result"
)

// checkerBinForTest returns the absolute path to the built checker, skipping the
// test when it has not been built yet.
func checkerBinForTest(t *testing.T) string {
	abs, err := filepath.Abs(filepath.Join("..", "..", "dist", "checker.exe"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(abs); err != nil {
		t.Skipf("checker not built at %s (run: make build)", abs)
	}
	return abs
}

func samplePath(t *testing.T, rel string) string {
	abs, err := filepath.Abs(filepath.Join("..", "..", "test-samples", rel))
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

// connect launches the server via `go run .` and returns a connected session.
func connect(t *testing.T) (*mcp.ClientSession, context.Context) {
	t.Helper()
	checker := checkerBinForTest(t)
	ctx := context.Background()

	cmd := exec.Command("go", "run", ".")
	cmd.Env = append(os.Environ(), "CHECKER_BIN="+checker)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0"}, nil)
	session, err := client.Connect(ctx, &mcp.CommandTransport{Command: cmd}, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session, ctx
}

func callCheck(t *testing.T, s *mcp.ClientSession, ctx context.Context, args map[string]any) result.CheckResult {
	t.Helper()
	res, err := s.CallTool(ctx, &mcp.CallToolParams{Name: "check_syntax", Arguments: args})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool returned error: %+v", res.Content)
	}
	b, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatalf("marshal structured content: %v", err)
	}
	var cr result.CheckResult
	if err := json.Unmarshal(b, &cr); err != nil {
		t.Fatalf("decode CheckResult: %v (raw %s)", err, b)
	}
	return cr
}

func TestCheckSyntaxValidJSON(t *testing.T) {
	s, ctx := connect(t)
	cr := callCheck(t, s, ctx, map[string]any{"file_path": samplePath(t, "json/config.json")})
	if !cr.Valid {
		t.Fatalf("expected valid, got %+v", cr)
	}
}

func TestCheckSyntaxDuplicateKeysStrict(t *testing.T) {
	s, ctx := connect(t)
	cr := callCheck(t, s, ctx, map[string]any{
		"file_path": samplePath(t, "json/duplicate_keys_not_correct.json"),
		"strict":    true,
	})
	if cr.Valid || len(cr.Errors) == 0 {
		t.Fatalf("expected duplicate-key errors, got %+v", cr)
	}
}

func TestCheckSyntaxInvalidPostgres(t *testing.T) {
	s, ctx := connect(t)
	cr := callCheck(t, s, ctx, map[string]any{
		"file_path": samplePath(t, "sql/postgres_query_not_correct.sql"),
		"type":      "sql:postgres",
	})
	if cr.Valid || len(cr.Errors) == 0 {
		t.Fatalf("expected postgres syntax error, got %+v", cr)
	}
}
