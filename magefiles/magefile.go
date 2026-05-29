// Build tasks for SyntaxChecker.
//
// Usage:
//
//	mage              # build checker + mcp into dist/ (default)
//	mage -l           # list all targets
//	mage <target>     # run a specific target
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Default target when running `mage` with no arguments.
var Default = Build

// version metadata injected via -ldflags into the checker binary.
func version() string   { return gitOr("dev", "describe", "--tags", "--always") }
func commit() string    { return gitOr("none", "rev-parse", "--short", "HEAD") }
func buildDate() string { return time.Now().UTC().Format("2006-01-02T15:04:05Z") }

// checkerLDFLAGS embeds version metadata; the mcp server has no version vars.
func checkerLDFLAGS() string {
	return fmt.Sprintf("-s -w -X main.version=%s -X main.commit=%s -X main.buildDate=%s",
		version(), commit(), buildDate())
}

const mcpLDFLAGS = "-s -w"

// Build compiles both binaries for the host platform into dist/.
func Build() {
	mg.Deps(Checker, Mcp)
}

// Dist is an alias for Build.
func Dist() { mg.Deps(Build) }

// Checker builds the syntax-checker CLI into dist/.
func Checker() error {
	return goBuild("", "", "./apps/checker", "syntax-checker", checkerLDFLAGS())
}

// Mcp builds the syntaxchecker-mcp server into dist/.
func Mcp() error {
	return goBuild("", "", "./apps/mcp-server", "syntaxchecker-mcp", mcpLDFLAGS)
}

// Windows cross-compiles both binaries for windows/amd64.
func Windows() error {
	if err := goBuild("windows", "amd64", "./apps/checker", "syntax-checker", checkerLDFLAGS()); err != nil {
		return err
	}
	return goBuild("windows", "amd64", "./apps/mcp-server", "syntaxchecker-mcp", mcpLDFLAGS)
}

// Linux cross-compiles both binaries for linux/amd64.
func Linux() error {
	if err := goBuild("linux", "amd64", "./apps/checker", "syntax-checker", checkerLDFLAGS()); err != nil {
		return err
	}
	return goBuild("linux", "amd64", "./apps/mcp-server", "syntaxchecker-mcp", mcpLDFLAGS)
}

// Test runs the test suite of every module.
func Test() error {
	for _, mod := range []string{"./apps/checker/...", "./apps/mcp-server/...", "./pkg/result/..."} {
		if err := sh.RunV("go", "test", mod); err != nil {
			return err
		}
	}
	return nil
}

// Lint runs `go vet` over every module.
func Lint() error {
	for _, mod := range []string{"./apps/checker/...", "./apps/mcp-server/...", "./pkg/result/..."} {
		if err := sh.RunV("go", "vet", mod); err != nil {
			return err
		}
	}
	return nil
}

// Installer builds the windows binaries and packages the Inno Setup installer.
// iscc must be on PATH (or set the ISCC env var to override the command).
func Installer() error {
	mg.Deps(Windows)
	iscc := os.Getenv("ISCC")
	if iscc == "" {
		iscc = "iscc"
	}
	return sh.RunV(iscc, "/DMyAppVersion="+version(), "installer.iss")
}

// Clean removes the dist/ directory.
func Clean() error {
	return os.RemoveAll("dist")
}

// goBuild compiles a package into dist/<name>(.exe) with CGO disabled.
// Empty goos/goarch builds for the host platform.
func goBuild(goos, goarch, pkg, name, ldflags string) error {
	out := filepath.Join("dist", name+exeExt(goos))
	env := map[string]string{"CGO_ENABLED": "0"}
	if goos != "" {
		env["GOOS"] = goos
	}
	if goarch != "" {
		env["GOARCH"] = goarch
	}
	fmt.Printf("→ building %s\n", out)
	return sh.RunWithV(env, "go", "build", "-trimpath", "-ldflags", ldflags, "-o", out, pkg)
}

// exeExt returns ".exe" when targeting Windows (or the host is Windows).
func exeExt(goos string) string {
	if goos == "windows" || (goos == "" && runtime.GOOS == "windows") {
		return ".exe"
	}
	return ""
}

// gitOr runs `git <args>` and returns the trimmed output, or fallback on error.
func gitOr(fallback string, args ...string) string {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return fallback
	}
	if s := strings.TrimSpace(string(out)); s != "" {
		return s
	}
	return fallback
}
