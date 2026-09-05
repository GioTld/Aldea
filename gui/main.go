package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend
var assets embed.FS

//go:embed icon.png
var icon []byte

func main() {
	userHome, err := os.UserHomeDir()
	if err != nil {
		userHome = "."
	}
	appDataDir := userHome + "/.aldea"

	app := NewApp(appDataDir)

	// Create Application Menu
	appMenu := menu.NewMenu()
	aldeaMenu := appMenu.AddSubmenu("Aldea")
	aldeaMenu.AddText("Mostrar Interfaz", keys.CmdOrCtrl("o"), func(_ *menu.CallbackData) {
		if app.ctx != nil {
			wailsRuntime.WindowShow(app.ctx)
		}
	})
	aldeaMenu.AddText("Pausar Servicio", nil, func(_ *menu.CallbackData) {
		app.PauseNode(true)
	})
	aldeaMenu.AddSeparator()
	aldeaMenu.AddText("Salir de Aldea", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		if app.ctx != nil {
			wailsRuntime.Quit(app.ctx)
		} else {
			os.Exit(0)
		}
	})

	err = wails.Run(&options.App{
		Title:             "Aldea — Distributed Storage Network",
		Width:             1140,
		Height:            760,
		HideWindowOnClose: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 7, G: 9, B: 11, A: 255},
		OnStartup:        app.startup,
		Menu:             appMenu,
		Linux: &linux.Options{
			Icon: icon,
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		fmt.Println("Error starting Aldea GUI:", err)
	}
}
