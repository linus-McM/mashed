//go:build dev

package terminal

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// R2: a `-tags dev` build also accepts the wails dev server origins.
func TestBridge_DevAcceptsLocalhost34115(t *testing.T) {
	ctx, b, _ := authBridge(t)
	for _, origin := range []string{"http://localhost:34115", "http://127.0.0.1:34115"} {
		h := authHeader(b)
		h.Set("Origin", origin)
		assert.NotEqual(t, http.StatusForbidden, getWS(t, ctx, b, "ghost", "", h), origin)
	}
}
