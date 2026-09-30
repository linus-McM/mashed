# Mashed — POSIX Makefile (thin wrapper over the primary `justfile`).
#
# Contributors who prefer `make` over `just` can use this. The canonical
# recipe set lives in `justfile`; this Makefile exposes only the handful
# of targets that CI and out-of-band contributors need.

.PHONY: help tools gen test build vet lint

help:
	@printf "Mashed Makefile targets:\n"
	@printf "  make tools   — install dev tooling (go-jsonschema for Story A codegen)\n"
	@printf "  make gen     — regenerate internal/uiadapter/uiast.gen.go from schemas/*.json\n"
	@printf "  make test    — run Go tests with -race\n"
	@printf "  make build   — build Go packages\n"
	@printf "  make vet     — run go vet\n"
	@printf "  make lint    — go vet + svelte-check\n"
	@printf "\nFull recipe list: just --list\n"

# Install developer tooling pinned via go.mod `tool` directive. Required by
# Story A TestCodegen_NoDrift, which auto-installs if missing but running
# this target up-front avoids the first-run latency hit.
tools:
	go install github.com/atombender/go-jsonschema@latest

# Regenerate internal/uiadapter/uiast.gen.go from schemas/*.json.
# CI asserts cleanliness via TestCodegen_NoDrift (AC-A.2).
gen:
	PATH="$$(go env GOPATH)/bin:$$PATH" go generate ./internal/uiadapter/...

test:
	go test -tags testing ./internal/... -race -count=1

build:
	go build ./...

vet:
	go vet ./...

lint: vet
	cd frontend && npm run check
