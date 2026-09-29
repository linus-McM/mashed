//go:build !dev

package terminal

// allowedOrigins is the Origin allowlist for the terminal WebSocket in
// production builds: the macOS Wails webview only (R2).
var allowedOrigins = map[string]bool{
	"wails://wails": true,
}
