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
	app := application.New(application.Options{
		Name:        "ScriptGuard",
		Description: "Python脚本监控与定时执行系统",
		Services: []application.Service{
			application.NewService(backend.NewApp()),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	mainWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "ScriptGuard - 脚本守护者",
		Width:  1400,
		Height: 900,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		Windows: application.WindowsWindow{
			DisableIcon: false,
		},
		BackgroundColour: application.NewRGB(15, 23, 42),
		URL:              "/",
		KeyBindings: map[string]func(window application.Window){
			"F12": func(window application.Window) {
				window.OpenDevTools()
			},
		},
	})
	mainWindow.Show()

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
