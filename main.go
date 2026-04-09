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

// resolveHelperPath finds the mashed-pty-helper binary next to the main
// executable (production) or under build/bin/ (development).
func resolveHelperPath() string {
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "mashed-pty-helper")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	if wd, err := os.Getwd(); err == nil {
		candidate := filepath.Join(wd, "build", "bin", "mashed-pty-helper")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	log.Printf("WARNING: mashed-pty-helper not found (terminal sessions will be unavailable)")
	return ""
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

	if helperPath != "" {
		sockPath := filepath.Join(os.TempDir(),
			fmt.Sprintf("mashed-pty-%d.sock", os.Getpid()))

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
				client, err := helper.Dial(sockPath)
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
