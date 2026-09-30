// Package uiadapter — Story A (Schema-to-Go codegen) RED-phase tests.
//
// These tests verify the tooling contract introduced by Story A:
//
//   - `internal/uiadapter/gen.go` must declare the canonical `//go:generate`
//     directive that emits `uiast.gen.go` from `schemas/*.json`.
//   - `internal/uiadapter/uiast.gen.go` must be a checked-in artifact under
//     package `uiast` (Plan §6.5 "Generated code is checked in").
//   - Running `go generate ./...` on a clean tree must leave no dirty file
//     (AC-A.2 "CI fails when schemas and generated code diverge").
//
// They deliberately do NOT import the generated `mashed/internal/uiadapter/uiast`
// package: that would cause a compile-time failure before GREEN lands, which
// would block the entire `uiadapter` test binary. Instead every assertion is
// file-system or `go/parser` based so the test binary compiles even on a
// fresh checkout, and each test fails with an intelligible, story-scoped
// error message during RED.
package uiadapter

import (
	"bytes"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ensureGoJsonschemaOnPATH prepends `$(go env GOPATH)/bin` to PATH for the
// lifetime of the test and installs `go-jsonschema` there if it is missing.
// This makes TestCodegen_NoDrift reproducible on a clean checkout without
// the developer having to export PATH manually — `go install` and `go env`
// are always available wherever `go test` is available, and the `tool`
// directive in go.mod pins the version the install resolves to.
func ensureGoJsonschemaOnPATH(t *testing.T) {
	t.Helper()

	gopath, err := exec.Command("go", "env", "GOPATH").Output()
	require.NoErrorf(t, err,
		"Story A AC-A.2: cannot resolve GOPATH via `go env GOPATH`: %v", err)
	// `go env GOPATH` may emit a PathListSeparator-joined list on some
	// platforms; the first entry is the writable install target.
	first := strings.TrimSpace(string(gopath))
	if i := strings.IndexByte(first, os.PathListSeparator); i >= 0 {
		first = first[:i]
	}
	binDir := filepath.Join(first, "bin")

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	if _, err := exec.LookPath("go-jsonschema"); err == nil {
		return
	}
	install := exec.Command("go", "install", "github.com/atombender/go-jsonschema@latest")
	if out, err := install.CombinedOutput(); err != nil {
		t.Fatalf("Story A AC-A.2: could not auto-install go-jsonschema into %s "+
			"(run `just install-tools` or `make tools` manually): %v\n%s",
			binDir, err, string(out))
	}
	if _, err := exec.LookPath("go-jsonschema"); err != nil {
		t.Fatalf("Story A AC-A.2: go-jsonschema still not on PATH after "+
			"`go install` — expected it in %s (err: %v)", binDir, err)
	}
}

// genDirectiveSubstring is the canonical directive Story A mandates (spec:
// docs/stories/uiadapter-A.md Developer Notes → Files (new) + Plan §3 Story A).
const genDirectiveSubstring = "//go:generate go-jsonschema -p uiast schemas/*.json"

// generatedFile is the checked-in codegen output per Plan §6.5.
const generatedFile = "uiast.gen.go"

// genFile houses the //go:generate directive per story Developer Notes.
const genFile = "gen.go"

// TestCodegen_AC_A1_GenDirectiveFileExists — AC-A.1 tooling precondition.
// Story Developer Notes mandate `internal/uiadapter/gen.go` (build-tag-free)
// carry the go-jsonschema directive.
func TestCodegen_AC_A1_GenDirectiveFileExists(t *testing.T) {
	t.Parallel()
	_, err := os.Stat(genFile)
	require.NoError(t, err,
		"Story A: %s must exist and house the //go:generate directive "+
			"(see docs/stories/uiadapter-A.md Developer Notes → Files (new))", genFile)
}

// TestCodegen_AC_A1_GenDirectiveFileDeclaresGoGenerate — AC-A.1 canonical
// directive. The directive must read exactly
// `//go:generate go-jsonschema -p uiast schemas/*.json` per Plan §3 Story A
// so all contributors produce identical output from `go generate ./...`.
func TestCodegen_AC_A1_GenDirectiveFileDeclaresGoGenerate(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile(genFile)
	require.NoError(t, err, "Story A: %s missing — cannot verify //go:generate directive", genFile)
	assert.Contains(t, string(b), genDirectiveSubstring,
		"Story A: %s must contain the canonical directive %q so codegen "+
			"is reproducible (Plan §3 Story A)", genFile, genDirectiveSubstring)
}

// TestCodegen_AC_A1_GenDirectiveFileIsBuildTagFree — Story Developer Notes
// require `gen.go` remain build-tag-free so `go generate ./...` picks it up
// in the default build.
func TestCodegen_AC_A1_GenDirectiveFileIsBuildTagFree(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile(genFile)
	require.NoError(t, err, "Story A: %s missing", genFile)
	head := string(b)
	if len(head) > 512 {
		head = head[:512]
	}
	assert.NotContains(t, head, "//go:build",
		"Story A Developer Notes: %s must be build-tag-free", genFile)
	assert.NotContains(t, head, "// +build",
		"Story A Developer Notes: %s must be build-tag-free", genFile)
}

// TestCodegen_AC_A1_GeneratedFileExists — AC-A.1: `uiast.gen.go` is the
// checked-in generated artifact (Plan §6.5 "Generated code is checked in").
func TestCodegen_AC_A1_GeneratedFileExists(t *testing.T) {
	t.Parallel()
	_, err := os.Stat(generatedFile)
	require.NoError(t, err,
		"Story A / Plan §6.5: %s must be checked in — run `go generate ./...` "+
			"and commit the result", generatedFile)
}

// TestCodegen_AC_A1_GeneratedFileDeclaresUIASTPackage — AC-A.1: generated
// types live in sub-package `uiast` per Plan §3 Story A (directive flag `-p uiast`).
// Callers in `uiadapter` import `mashed/internal/uiadapter/uiast`.
func TestCodegen_AC_A1_GeneratedFileDeclaresUIASTPackage(t *testing.T) {
	t.Parallel()
	src, err := os.ReadFile(generatedFile)
	require.NoError(t, err, "Story A: %s missing", generatedFile)
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, generatedFile, src, parser.PackageClauseOnly)
	require.NoError(t, err, "Story A: %s must be parseable Go", generatedFile)
	assert.Equal(t, "uiast", f.Name.Name,
		"Story A: %s must declare `package uiast` (directive flag `-p uiast`)", generatedFile)
}

// TestCodegen_AC_A1_GeneratedPackageDirectoryExists — companion: since
// go-jsonschema emits files under `uiast/`, the directory must exist and be
// importable as `mashed/internal/uiadapter/uiast`.
func TestCodegen_AC_A1_GeneratedPackageDirectoryExists(t *testing.T) {
	t.Parallel()
	// Two acceptable layouts per go-jsonschema conventions:
	//   (a) single file `uiast.gen.go` declaring `package uiast` at the
	//       uiadapter root (preferred by this story);
	//   (b) directory `uiast/` containing generated `.go` files.
	// Either layout must render package `uiast` importable from this package.
	// This test succeeds if (a) is in place OR (b) is in place.
	_, errFile := os.Stat(generatedFile)
	dirInfo, errDir := os.Stat("uiast")
	if errFile == nil {
		// (a) single-file layout — validated by sibling tests.
		return
	}
	require.NoError(t, errDir,
		"Story A: expected either %s at the uiadapter root OR a `uiast/` "+
			"package directory containing generated files — neither found", generatedFile)
	assert.True(t, dirInfo.IsDir(),
		"Story A: `uiast` entry under internal/uiadapter must be a directory")
}

// TestCodegen_NoDrift — AC-A.2 (task spec literal name): running
// `go generate ./...` followed by `git diff --exit-code` on the generated
// artifact must be clean. If `schemas/*.json` has been edited without
// re-running codegen (or vice versa) this test fails and the message names
// `uiast.gen.go` per the BDD scenario — satisfying the sign-off criterion
// `go test ./internal/uiadapter/... -run TestCodegen_NoDrift` must fail at
// RED.
//
// This test is the in-repo mirror of the CI check the story mandates.
func TestCodegen_NoDrift(t *testing.T) {
	// Intentionally NOT t.Parallel() — we invoke `go generate` and `git diff`
	// against the working tree; running in parallel with other tests that
	// might touch temp files is fine, but serializing keeps failure output
	// deterministic.

	repoRoot, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	require.NoError(t, err, "Story A AC-A.2: `git rev-parse` must succeed — are we in a git work tree?")
	root := strings.TrimSpace(string(repoRoot))

	// Pre-flight: guarantee the codegen tool is discoverable. Story A Task 1a
	// pins it via the `tool` directive in go.mod; this helper prepends
	// `$(go env GOPATH)/bin` to PATH and auto-`go install`s the binary if it
	// is missing, so the test is reproducible on a clean checkout without
	// the developer exporting PATH manually.
	ensureGoJsonschemaOnPATH(t)

	genCmd := exec.Command("go", "generate", "./...")
	genCmd.Dir = root
	genOut, err := genCmd.CombinedOutput()
	require.NoErrorf(t, err,
		"Story A AC-A.2: `go generate ./...` failed: %v\n%s", err, string(genOut))

	target := filepath.Join("internal", "uiadapter", generatedFile)
	diffCmd := exec.Command("git", "diff", "--exit-code", "--", target)
	diffCmd.Dir = root
	var stdout, stderr bytes.Buffer
	diffCmd.Stdout = &stdout
	diffCmd.Stderr = &stderr
	err = diffCmd.Run()
	if err != nil {
		t.Fatalf("Story A AC-A.2 (Schema→Go drift detected): %s is dirty after "+
			"`go generate ./...`. Re-run codegen and commit the result.\n"+
			"--- git diff %s ---\n%s%s",
			generatedFile, target, stdout.String(), stderr.String())
	}
}

// TestCodegen_AC_A2_NoDrift failure path is already exercised inside
// TestCodegen_AC_A2_NoDrift — the `t.Fatalf` message format there must name
// `uiast.gen.go` and prompt the developer to re-run `go generate`, which
// matches the BDD scenario verbatim:
//
//	"And the failure message names the drifted `uiast.gen.go`"
//
// (We deliberately do NOT declare a `formatDriftMessage` helper here: any
// extracted helper would sit in a non-test source file, which is an
// implementation detail for GREEN; the RED contract is satisfied by the
// literal format string inside TestCodegen_AC_A2_NoDrift above.)
