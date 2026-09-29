package main

import (
	"fmt"
	"strings"
	"testing"
	"testing/fstest"
)

// R21: a binary built without the frontend (placeholder dist only) logs a
// clear warning at startup instead of silently serving a blank window.
func TestWarnIfDistMissing(t *testing.T) {
	var logged []string
	logf := func(format string, args ...any) { logged = append(logged, fmt.Sprintf(format, args...)) }

	warnIfDistMissing(fstest.MapFS{"frontend/dist/.gitkeep": {}}, logf)
	if len(logged) != 1 || !strings.Contains(logged[0], "frontend/dist/index.html") {
		t.Fatalf("missing index.html: logged %q, want one warning naming frontend/dist/index.html", logged)
	}

	logged = nil
	warnIfDistMissing(fstest.MapFS{"frontend/dist/index.html": {Data: []byte("<html>")}}, logf)
	if len(logged) != 0 {
		t.Fatalf("index.html present: logged %q, want nothing", logged)
	}
}
