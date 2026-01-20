package main

import (
	"embed"
	"log"
	"sync/atomic"

	"scriptguard/backend"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

func main() {
	// 控制是否允许真正退出（只有托盘菜单"退出"才允许）
	var allowQuit atomic.Bool

	app := application.New(application.Options{
		Name:        "ScriptGuard",
		Description: "Python脚本监控与定时执行系统",
		Services: []application.Service{
			application.NewService(backend.NewApp()),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Icon: appIcon,
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false, // Mac 也支持托盘
		},
	})

	mainWindow := app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
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
		KeyBindings: map[string]func(window *application.WebviewWindow){
			"F12": func(window *application.WebviewWindow) {
				window.OpenDevTools()
			},
		},
	})

	// 显示窗口函数
	showWindow := func() {
		mainWindow.Show()
		mainWindow.Focus()
	}

	// 切换窗口显示/隐藏
	toggleWindow := func() {
		if mainWindow.IsVisible() {
			mainWindow.Hide()
			return
		}
		showWindow()
	}

	// 拦截窗口关闭事件：最小化到托盘而非退出
	mainWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		if allowQuit.Load() {
			return // 允许退出
		}
		mainWindow.Hide()
		e.Cancel() // 取消关闭，改为隐藏
	})

	// 创建系统托盘
	systemTray := app.NewSystemTray()
	systemTray.SetIcon(appIcon)
	systemTray.SetTooltip("ScriptGuard - 脚本守护者")

	// 托盘右键菜单
	trayMenu := app.NewMenu()
	trayMenu.Add("显示/隐藏窗口").OnClick(func(_ *application.Context) {
		toggleWindow()
	})
	trayMenu.AddSeparator()
	trayMenu.Add("退出应用").OnClick(func(_ *application.Context) {
		allowQuit.Store(true)
		app.Quit()
	})
	systemTray.SetMenu(trayMenu)

	// 托盘交互
	systemTray.OnRightClick(func() {
		systemTray.OpenMenu()
	})
	systemTray.OnDoubleClick(func() {
		showWindow()
	})
	systemTray.OnClick(func() {
		// 单击也显示窗口
		showWindow()
	})

	mainWindow.Show()

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
