// Package uiadapter — Story A (Schema-to-Go codegen) RED-phase tests.
//
// These tests assert the SHAPE CONTRACT between the hand-written UIAST types
// in `internal/uiadapter/schema.go` and the generated types the Story A
// codegen will emit into `internal/uiadapter/uiast.gen.go` (package `uiast`).
//
// Story A AC-A.1 permits two outcomes:
//  1. delete the hand-written types and have callers use the generated ones, OR
//  2. keep both and declare a compile-time shape assertion
//     (`var _ ManualUIAST = GeneratedUIAST{}`) so the two cannot drift.
//
// During RED, neither the generated file nor the co-existence assertion
// exists. Rather than import `mashed/internal/uiadapter/uiast` (which would
// compile-break the whole test binary), these tests PARSE the generated
// source file with `go/parser` and compare JSON tags against the hand-written
// struct tags discovered via `reflect`. The tests fail with clear,
// story-scoped error messages until GREEN lands.
package uiadapter

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parsedGenFile loads uiast.gen.go into an *ast.File for shape inspection.
// Tests that need generated-type data call this helper; on missing file the
// helper t.Fatals with a story-scoped message so every shape test produces
// the same legible failure during RED.
func parsedGenFile(t *testing.T) *ast.File {
	t.Helper()
	src, err := os.ReadFile(generatedFile)
	require.NoErrorf(t, err,
		"Story A AC-A.1: %s missing — run `go generate ./...` to produce "+
			"the checked-in artifact", generatedFile)
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, generatedFile, src, parser.AllErrors)
	require.NoErrorf(t, err, "Story A: %s must be parseable Go: %v", generatedFile, err)
	return f
}

// findStruct returns the *ast.StructType for a top-level type declaration
// named `name`, or nil if not found.
func findStruct(f *ast.File, name string) *ast.StructType {
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || ts.Name == nil || ts.Name.Name != name {
				continue
			}
			if st, ok := ts.Type.(*ast.StructType); ok {
				return st
			}
		}
	}
	return nil
}

// jsonTagsFromStruct extracts the JSON key for every field of a parsed
// struct, preserving `omitempty` semantics as `name,omitempty`. Fields with
// JSON tag `"-"` are skipped (they never reach the wire).
func jsonTagsFromStruct(st *ast.StructType) []string {
	var out []string
	for _, field := range st.Fields.List {
		if field.Tag == nil {
			continue
		}
		raw, err := strconv.Unquote(field.Tag.Value)
		if err != nil {
			continue
		}
		tag := reflect.StructTag(raw).Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		out = append(out, tag)
	}
	sort.Strings(out)
	return out
}

// jsonTagsFromReflect extracts the JSON key (with omitempty) for every
// exported field of the reflect.Type, matching the format produced by
// jsonTagsFromStruct so sets can be compared directly.
func jsonTagsFromReflect(rt reflect.Type) []string {
	var out []string
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		tag := f.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		out = append(out, tag)
	}
	sort.Strings(out)
	return out
}

// jsonKeyOnly strips an optional ",omitempty" suffix.
func jsonKeyOnly(tag string) string {
	if i := strings.Index(tag, ","); i >= 0 {
		return tag[:i]
	}
	return tag
}

// TestGeneratedTypes_ShapeAssertion — task spec literal name: the single
// entry point that discharges AC-A.1 "generated types match hand-written
// UIAST types." Delegates to the per-struct shape tests below for
// granular diagnostics. Sign-off criterion:
//
//	go test ./internal/uiadapter/... -run TestGeneratedTypes_ShapeAssertion
//
// must FAIL during RED because `uiast.gen.go` does not exist.
//
// Note: we intentionally emulate a *compile-time* assertion at runtime via
// `go/parser` + `reflect`. A literal `var _ uiadapter.UIAST = uiast.UIAST{}`
// assertion would compile-break the test binary before the generated
// package ships, which would hide every other failure in the package and
// block parallel work on Stories B / C.
func TestGeneratedTypes_ShapeAssertion(t *testing.T) {
	t.Parallel()
	f := parsedGenFile(t) // fails RED with "uiast.gen.go missing"

	// Every hand-written type that must have a matching generated counterpart.
	// This list is the canonical replacement target the GREEN implementer
	// must deliver (AC-A.1 option 1: delete hand-written types, or option 2:
	// add `var _ Old = New{}` co-existence assertions).
	type shapeCase struct {
		handWritten reflect.Type
		candidates  []string // generated names accepted, first match wins
		rationale   string
	}
	cases := []shapeCase{
		{reflect.TypeOf(UIAST{}), []string{"UIAST"},
			"§3.1 envelope — root of every translator payload"},
		{reflect.TypeOf(UINode{}), []string{"UINode", "Node", "UiastNode"},
			"§3.2 discriminated-union node variant"},
		{reflect.TypeOf(WidgetNode{}), []string{"WidgetNode", "Widget"},
			"§3.2 + §4.7.1 widget control metadata (camelCase `repoRootRelative`, `maxLength`)"},
		{reflect.TypeOf(WidgetOption{}), []string{"WidgetOption", "Option"},
			"§3.2 option entry inside widget nodes"},
		{reflect.TypeOf(Diagnostics{}), []string{"Diagnostics"},
			"§3.1 + §4.3 diagnostics — CancelReason must keep JSON tag \"-\""},
	}

	for _, tc := range cases {
		var st *ast.StructType
		var chosen string
		for _, name := range tc.candidates {
			if s := findStruct(f, name); s != nil {
				st, chosen = s, name
				break
			}
		}
		require.NotNilf(t, st,
			"Story A AC-A.1: %s must declare a struct matching %s "+
				"(tried %v) — %s",
			generatedFile, tc.handWritten.Name(), tc.candidates, tc.rationale)

		want := jsonTagsFromReflect(tc.handWritten)
		got := jsonTagsFromStruct(st)
		assert.Equalf(t, want, got,
			"Story A AC-A.1 shape drift: generated %s JSON tags differ "+
				"from hand-written %s — %s",
			chosen, tc.handWritten.Name(), tc.rationale)
	}
}

// TestUIASTShape_AC_A1_GeneratedUIASTMatchesHandWritten — AC-A.1: the
// generated `uiast.UIAST` must carry the same JSON tags as the hand-written
// `uiadapter.UIAST`, including identical `omitempty` placement, so a shape
// assertion (`var _ UIAST = uiast.UIAST{}`) would hold at compile time.
func TestUIASTShape_AC_A1_GeneratedUIASTMatchesHandWritten(t *testing.T) {
	t.Parallel()
	f := parsedGenFile(t)
	st := findStruct(f, "UIAST")
	require.NotNilf(t, st,
		"Story A AC-A.1: %s must declare `type UIAST struct { ... }`", generatedFile)

	want := jsonTagsFromReflect(reflect.TypeOf(UIAST{}))
	got := jsonTagsFromStruct(st)
	assert.Equal(t, want, got,
		"Story A AC-A.1: generated UIAST JSON tags must match hand-written "+
			"uiadapter.UIAST (envelope §3.1). Drift means Ollama `format:` "+
			"and Claude `input_schema` would render different shapes.")
}

// TestUIASTShape_AC_A1_GeneratedUINodeMatchesHandWritten — AC-A.1 for the
// discriminated-union node variant (§3.2).
func TestUIASTShape_AC_A1_GeneratedUINodeMatchesHandWritten(t *testing.T) {
	t.Parallel()
	f := parsedGenFile(t)
	// Accept either `UINode` (matching hand-written) or go-jsonschema's
	// default `Node` / `UiastNode` auto-name. The story permits rename as
	// long as shape matches; expose whichever the generator produces.
	var st *ast.StructType
	var chosen string
	for _, candidate := range []string{"UINode", "Node", "UiastNode"} {
		if s := findStruct(f, candidate); s != nil {
			st, chosen = s, candidate
			break
		}
	}
	require.NotNilf(t, st,
		"Story A AC-A.1: %s must declare a UINode-equivalent struct "+
			"(tried UINode, Node, UiastNode)", generatedFile)

	want := jsonTagsFromReflect(reflect.TypeOf(UINode{}))
	got := jsonTagsFromStruct(st)
	assert.Equalf(t, want, got,
		"Story A AC-A.1: generated %s JSON tags must match hand-written "+
			"uiadapter.UINode (§3.2)", chosen)
}

// TestUIASTShape_AC_A1_GeneratedWidgetMatchesHandWritten — AC-A.1 for the
// widget-control metadata (§3.2 + §4.7.1).
func TestUIASTShape_AC_A1_GeneratedWidgetMatchesHandWritten(t *testing.T) {
	t.Parallel()
	f := parsedGenFile(t)
	var st *ast.StructType
	var chosen string
	for _, candidate := range []string{"WidgetNode", "Widget"} {
		if s := findStruct(f, candidate); s != nil {
			st, chosen = s, candidate
			break
		}
	}
	require.NotNilf(t, st,
		"Story A AC-A.1: %s must declare a widget struct (tried WidgetNode, Widget)",
		generatedFile)

	want := jsonTagsFromReflect(reflect.TypeOf(WidgetNode{}))
	got := jsonTagsFromStruct(st)
	assert.Equalf(t, want, got,
		"Story A AC-A.1: generated %s JSON tags must match hand-written "+
			"uiadapter.WidgetNode — the camelCase `repoRootRelative` and "+
			"`maxLength` keys are load-bearing (§4.7.1 strict decoding)",
		chosen)
}

// TestUIASTShape_AC_A1_GeneratedWidgetOptionMatchesHandWritten — AC-A.1
// for the option entry inside widget nodes.
func TestUIASTShape_AC_A1_GeneratedWidgetOptionMatchesHandWritten(t *testing.T) {
	t.Parallel()
	f := parsedGenFile(t)
	var st *ast.StructType
	var chosen string
	for _, candidate := range []string{"WidgetOption", "Option"} {
		if s := findStruct(f, candidate); s != nil {
			st, chosen = s, candidate
			break
		}
	}
	require.NotNilf(t, st,
		"Story A AC-A.1: %s must declare a widget option struct "+
			"(tried WidgetOption, Option)", generatedFile)

	want := jsonTagsFromReflect(reflect.TypeOf(WidgetOption{}))
	got := jsonTagsFromStruct(st)
	assert.Equalf(t, want, got,
		"Story A AC-A.1: generated %s JSON tags must match hand-written "+
			"uiadapter.WidgetOption", chosen)
}

// TestUIASTShape_AC_A1_GeneratedDiagnosticsMatchesHandWritten — AC-A.1
// for Diagnostics. CancelReason has JSON tag `"-"` in the hand-written
// type; the generated equivalent must either omit the field entirely OR
// preserve the tag `"-"` so it never reaches the wire (§4.3).
func TestUIASTShape_AC_A1_GeneratedDiagnosticsMatchesHandWritten(t *testing.T) {
	t.Parallel()
	f := parsedGenFile(t)
	st := findStruct(f, "Diagnostics")
	require.NotNilf(t, st,
		"Story A AC-A.1: %s must declare `type Diagnostics struct { ... }`",
		generatedFile)

	want := jsonTagsFromReflect(reflect.TypeOf(Diagnostics{}))
	got := jsonTagsFromStruct(st)
	assert.Equalf(t, want, got,
		"Story A AC-A.1: generated Diagnostics JSON tags must match "+
			"hand-written uiadapter.Diagnostics")

	// Extra: the generator must not leak a `cancel_reason` key onto the
	// wire. Accept either "no CancelReason field" or "JSON tag -".
	for _, field := range st.Fields.List {
		for _, nm := range field.Names {
			if nm.Name != "CancelReason" {
				continue
			}
			if field.Tag == nil {
				t.Fatalf("Story A AC-A.1 §4.3: generated Diagnostics.CancelReason " +
					"has no struct tag — must be `json:\"-\"` to stay off the wire")
			}
			raw, err := strconv.Unquote(field.Tag.Value)
			require.NoError(t, err)
			tag := reflect.StructTag(raw).Get("json")
			assert.Equal(t, "-", tag,
				"Story A AC-A.1 §4.3: CancelReason must carry JSON tag \"-\"")
		}
	}
}

// TestUIASTShape_AC_A1_EnvelopeKeysPresentInGenerated — cross-check against
// the canonical schema: every envelope key the schema declares must appear
// as a JSON tag on the generated UIAST struct. Catches the case where tags
// technically compare equal but use the wrong casing.
func TestUIASTShape_AC_A1_EnvelopeKeysPresentInGenerated(t *testing.T) {
	t.Parallel()
	f := parsedGenFile(t)
	st := findStruct(f, "UIAST")
	require.NotNilf(t, st, "Story A AC-A.1: %s must declare UIAST", generatedFile)

	got := map[string]struct{}{}
	for _, tag := range jsonTagsFromStruct(st) {
		got[jsonKeyOnly(tag)] = struct{}{}
	}
	for _, key := range uiastEnvelopeKeys {
		_, present := got[key]
		assert.Truef(t, present,
			"Story A §3.1: generated UIAST missing JSON key %q", key)
	}
}
