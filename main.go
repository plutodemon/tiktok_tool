package main

import (
	_ "embed"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"tiktok_tool/capture"
	"tiktok_tool/config"
	"tiktok_tool/lkit"
	"tiktok_tool/llog"
	"tiktok_tool/ui"
)

//go:embed tiktok.png
var iconBytes []byte

func main() {
	lkit.InitCrashLog()
	defer lkit.CrashLog()

	if err := config.LoadConfig(); err != nil {
		panic(fmt.Sprintf("初始化配置失败: %v", err))
	}

	if err := llog.Init(config.GetConfig().LogConfig); err != nil {
		panic(fmt.Sprintf("初始化日志系统失败: %v", err))
	}
	defer llog.Cleanup()

	// 设置全局panic处理
	defer llog.HandlePanic()

	myApp := app.New()
	myApp.SetIcon(&fyne.StaticResource{
		StaticName:    "icon",
		StaticContent: iconBytes,
	})
	window := myApp.NewWindow("抖音直播推流配置抓取")
	window.Resize(fyne.NewSize(600, 300))
	window.SetFixedSize(true)
	window.SetMaster()
	window.CenterOnScreen()

	if !capture.CheckNpcapInstalled() {
		ui.ShowInstallDialog(window)
		window.ShowAndRun()
		return
	}

	ui.NewMainWindow(myApp, window)
	window.ShowAndRun()
}
