package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func (w *SettingsWindow) createWindowTab() fyne.CanvasObject {
	winContainer := container.NewVBox(
		w.minimizeOnClose,
		w.openLiveWhenStart,
	)

	winHelp := widget.NewRichTextFromMarkdown("### 窗口说明\n\n" +
		"* **关闭窗口时最小化到托盘**：勾选后关闭窗口时程序将最小化到系统托盘而不退出" +
		"\n* **启动时打开直播**：勾选后程序启动时自动打开直播窗口\n\n")

	return container.NewVBox(
		winContainer,
		layout.NewSpacer(),
		winHelp,
	)
}
