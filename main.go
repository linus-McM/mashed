package main

import (
	"embed"
	"log"

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
		if _, err := app.TakeScreenshot(); err != nil {
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

func main() {
	app := NewApp()

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
