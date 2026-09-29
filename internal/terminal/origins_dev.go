//go:build dev

package terminal

// allowedOrigins for `wails dev` builds also accepts the Vite dev server that
// serves the frontend in a browser (R2).
var allowedOrigins = map[string]bool{
	"wails://wails":          true,
	"http://localhost:34115": true,
	"http://127.0.0.1:34115": true,
}
