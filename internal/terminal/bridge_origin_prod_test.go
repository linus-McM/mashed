//go:build !dev

package terminal

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// R2: a production build rejects the wails dev server origin.
func TestBridge_ProdRejectsDevOrigin(t *testing.T) {
	ctx, b, _ := authBridge(t)
	h := authHeader(b)
	h.Set("Origin", "http://localhost:34115")
	assert.Equal(t, http.StatusForbidden, getWS(t, ctx, b, "bmad-x", "", h))
}
