package main

import (
	_ "embed"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"tiktok_tool/capture"
	"tiktok_tool/config"
	"tiktok_tool/ui"
)

//go:embed tiktok.png
var iconBytes []byte

func main() {
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

	// 加载配置文件
	if err := config.LoadSettings(); err != nil {
		dialog.ShowError(err, window)
	}

	if !capture.CheckNpcapInstalled() {
		ui.ShowInstallDialog(window)
		window.ShowAndRun()
		return
	}

	ui.NewMainWindow(myApp, window)
	window.ShowAndRun()
}
