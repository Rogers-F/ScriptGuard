package main

import (
	"embed"
	"log"
	"scriptguard/backend"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	appInstance := backend.NewApp()

	app := application.New(application.Options{
		Name:        "ScriptGuard",
		Description: "Python脚本监控与定时执行系统",
		Assets: application.AssetOptions{
			FS: assets,
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
		Bind: []interface{}{
			appInstance,
		},
	})

	window := app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
		Title:  "ScriptGuard - 脚本守护者",
		Width:  1400,
		Height: 900,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(15, 23, 42),
		URL:              "/",
	})

	window.On(application.EventWindowDidLoad, func(event *application.WindowEvent) {
		appInstance.Startup(event.Context())
	})

	window.On(application.EventWindowClosing, func(event *application.WindowEvent) {
		appInstance.Shutdown()
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
