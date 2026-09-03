package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	userHome, err := os.UserHomeDir()
	if err != nil {
		userHome = "."
	}
	appDataDir := userHome + "/.aldea"

	app := NewApp(appDataDir)

	err = wails.Run(&options.App{
		Title:  "Aldea — Distributed Storage Network",
		Width:  1140,
		Height: 760,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 7, G: 9, B: 11, A: 255},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		fmt.Println("Error starting Aldea GUI:", err)
	}
}
