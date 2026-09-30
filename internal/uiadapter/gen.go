// Package uiadapter — Story A canonical Schema→Go codegen directive.
//
// The canonical invocation documented in Plan §3 Story A is:
//     //go:generate go-jsonschema -p uiast schemas/*.json
// `go generate` does not expand shell globs itself, so the effective
// directive below wraps the call in `sh -c` to let the shell expand
// `schemas/*.json`. Contributors editing schemas run:
//     go generate ./internal/uiadapter/...
// CI asserts the result is clean via AC-A.2 TestCodegen_NoDrift.
//
//go:generate sh -c "go-jsonschema -p uiast schemas/*.json --only-models --disable-omitzero --tags json -t -o uiast.gen.go"
package uiadapter
