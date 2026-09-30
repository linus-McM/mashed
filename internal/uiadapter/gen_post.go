// Package uiadapter — post-processing step for Story A codegen.
//
// go-jsonschema does not emit a build constraint header, so we prepend
// one after generation. The generated file declares `package uiast`
// but lives at the uiadapter package root (per tests); the build tag
// below excludes it from compilation so the two packages coexist
// without a "mixed packages" build error while remaining parseable
// via go/parser for AST-based shape assertions (AC-A.1).
//
// This directive runs AFTER the one in gen.go because `go generate`
// processes files in alphabetical order (gen.go < gen_post.go) and
// directives within a file in top-to-bottom order.
package uiadapter

//go:generate sh -c "printf '%s\\n\\n' '//go:build ignore' | cat - uiast.gen.go > uiast.gen.go.tmp && mv uiast.gen.go.tmp uiast.gen.go"
