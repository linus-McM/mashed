package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"mashed/internal/terminal/helper"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

const repoURL = "https://github.com/user/mashed"

// buildMenu constructs the native menu bar for the application.
func buildMenu(app *App) *menu.Menu {
	appMenu := menu.NewMenu()

	mashed := appMenu.AddSubmenu("mashed")
	mashed.AddText("About mashed", nil, func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:about")
	})
	mashed.AddSeparator()
	mashed.AddText("Settings...", keys.CmdOrCtrl(","), func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:navigate", "settings")
	})
	mashed.AddSeparator()
	mashed.AddText("Quit mashed", keys.CmdOrCtrl("q"), func(cd *menu.CallbackData) {
		runtime.Quit(app.ctx)
	})

	file := appMenu.AddSubmenu("File")
	file.AddText("New Agent", keys.CmdOrCtrl("n"), func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:navigate", "spawn")
	})
	file.AddText("New Repository...", keys.Combo("n", keys.CmdOrCtrlKey, keys.ShiftKey), func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:navigate", "new-repo")
	})
	file.AddSeparator()
	file.AddText("Open Workspace...", keys.CmdOrCtrl("o"), func(cd *menu.CallbackData) {
		path, _ := app.PickDirectory()
		if path != "" {
			runtime.EventsEmit(app.ctx, "menu:open-workspace", path)
		}
	})
	file.AddSeparator()
	file.AddText("Take Screenshot", keys.Combo("s", keys.CmdOrCtrlKey, keys.ShiftKey), func(cd *menu.CallbackData) {
		app.mu.Lock()
		rp := app.activeRepoPath
		pt := app.activePaneTarget
		app.mu.Unlock()
		if rp == "" {
			log.Println("screenshot: no active context, skipping")
			return
		}
		if _, err := app.TakeScreenshot(rp, pt); err != nil {
			log.Printf("screenshot failed: %v", err)
		}
	})

	appMenu.Append(menu.EditMenu())

	view := appMenu.AddSubmenu("View")
	view.AddText("Feed", keys.CmdOrCtrl("1"), func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:navigate", "feed")
	})
	view.AddText("Workflows", keys.CmdOrCtrl("2"), func(cd *menu.CallbackData) {
		runtime.EventsEmit(app.ctx, "menu:navigate", "workflows")
	})
	view.AddSeparator()
	view.AddText("Minimize", keys.CmdOrCtrl("m"), func(cd *menu.CallbackData) {
		runtime.WindowMinimise(app.ctx)
	})
	view.AddText("Toggle Fullscreen", keys.Combo("f", keys.ControlKey, keys.CmdOrCtrlKey), func(cd *menu.CallbackData) {
		runtime.WindowToggleMaximise(app.ctx)
	})

	help := appMenu.AddSubmenu("Help")
	help.AddText("mashed on GitHub", nil, func(cd *menu.CallbackData) {
		runtime.BrowserOpenURL(app.ctx, repoURL)
	})

	return appMenu
}

// resolveHelperPath finds the mashed-pty-helper binary. Development path
// (build/bin/) is checked first so the freshly-signed binary is preferred
// over a potentially stale copy inside the .app bundle.
func resolveHelperPath() string {
	if wd, err := os.Getwd(); err == nil {
		candidate := filepath.Join(wd, "build", "bin", "mashed-pty-helper")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "mashed-pty-helper")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	log.Printf("WARNING: mashed-pty-helper not found (terminal sessions will be unavailable)")
	return ""
}

// setupHelperSocketDir creates a private (0700) directory for the PTY helper
// socket and returns a cleanup func that removes it (R12).
func setupHelperSocketDir() (string, func(), error) {
	dir, err := os.MkdirTemp("", "mashed-pty-")
	if err != nil {
		return "", func() {}, err
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		os.RemoveAll(dir)
		return "", func() {}, err
	}
	return dir, func() { os.RemoveAll(dir) }, nil
}

// dialHelperSecure dials the helper only if its socket is owner-only (R12).
func dialHelperSecure(path string) (*helper.Client, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if fi.Mode()&os.ModeSocket == 0 {
		return nil, fmt.Errorf("pty helper: %s is not a socket", path)
	}
	if perm := fi.Mode().Perm(); perm&0o077 != 0 {
		return nil, fmt.Errorf("pty helper: socket %s is accessible to other users (mode %o)", path, perm)
	}
	return helper.Dial(path)
}

// waitForSocket polls until the Unix socket file appears or the timeout expires.
func waitForSocket(path string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

func main() {
	// Resolve and launch PTY helper.
	helperPath := resolveHelperPath()
	var helperClient *helper.Client
	var helperCmd *exec.Cmd

	sockDir, cleanupSockDir, sockDirErr := "", func() {}, error(nil)
	if helperPath != "" {
		sockDir, cleanupSockDir, sockDirErr = setupHelperSocketDir()
		if sockDirErr != nil {
			log.Printf("WARNING: PTY helper socket dir: %v", sockDirErr)
			helperPath = ""
		}
	}
	defer cleanupSockDir()

	if helperPath != "" {
		sockPath := filepath.Join(sockDir, "pty.sock")

		helperCmd = exec.Command(helperPath)
		helperCmd.Env = append(os.Environ(),
			"MASHED_PTY_SOCK="+sockPath,
			fmt.Sprintf("MASHED_PARENT_PID=%d", os.Getpid()),
		)
		helperCmd.Stdout = os.Stdout
		helperCmd.Stderr = os.Stderr

		if err := helperCmd.Start(); err != nil {
			log.Printf("WARNING: PTY helper failed to start: %v", err)
		} else {
			if waitForSocket(sockPath, 3*time.Second) {
				client, err := dialHelperSecure(sockPath)
				if err != nil {
					log.Printf("WARNING: PTY helper dial failed: %v", err)
				} else {
					helperClient = client
				}
			} else {
				log.Printf("WARNING: PTY helper did not become ready in time")
			}
		}
	}

	app := NewApp(helperClient)

	// Story uiadapter-logging-1: ensure ./logs exists before the
	// production logger tries to open the daily file. Failure is
	// non-fatal — NewProductionLogger will degrade to stdout-only.
	if err := os.MkdirAll("./logs", 0o755); err != nil {
		log.Printf("uiadapter: log dir create failed: %v", err)
	}

	// Graceful shutdown: close client, signal helper, wait for exit.
	defer func() {
		if helperClient != nil {
			helperClient.Close()
		}
		if helperCmd != nil && helperCmd.Process != nil {
			helperCmd.Process.Signal(syscall.SIGTERM)
			helperCmd.Wait()
		}
	}()

	err := wails.Run(&options.App{
		Title:            "Mashed",
		Width:            1280,
		Height:           800,
		MinWidth:         800,
		MinHeight:        600,
		DisableResize:    false,
		Frameless:        true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Menu:             buildMenu(app),
		// Design system: --bg-deepest #07080a
		BackgroundColour: &options.RGBA{R: 7, G: 8, B: 10, A: 255},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			TitleBar: mac.TitleBarHiddenInset(),
			WebviewIsTransparent: true,
			WindowIsTranslucent:  false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
