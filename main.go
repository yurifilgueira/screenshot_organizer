package main

import (
	"embed"
	_ "embed"
	"fmt"
	"log"

	"fyne.io/systray"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:build/windows/icon.ico
var iconData []byte

//go:embed all:frontend/dist
var assets embed.FS

func main() {

	fmt.Println("Icon Data: ", iconData)
	fmt.Println("Starting Screenshot Organizer with Tray Menu")

	app := NewApp()

	go func() {
		systray.Run(func() {
			systray.SetIcon(iconData)
			systray.SetTooltip("Screenshot Organizer")

			mShow := systray.AddMenuItem("Show App", "Show the main window")
			systray.AddSeparator()
			mQuit := systray.AddMenuItem("Quit", "Quit the application")

			for {
				select {
				case <-mShow.ClickedCh:
					if app.ctx != nil {
						runtime.WindowShow(app.ctx)
					}
				case <-mQuit.ClickedCh:
					if app.ctx != nil {
						app.hideOnClose = false
						runtime.Quit(app.ctx)
					} else {
						systray.Quit()
					}
				}
			}
		}, func() {
		})
	}()

	err := wails.Run(&options.App{
		Title:  "Screenshot Organizer",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnBeforeClose:    app.beforeClose,
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
