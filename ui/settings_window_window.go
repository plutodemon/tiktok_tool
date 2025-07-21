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
	)

	winHelp := widget.NewRichTextFromMarkdown("### 窗口说明\n\n" +
		"* **关闭窗口时最小化到托盘**：勾选后关闭窗口时程序将最小化到系统托盘而不退出")

	return container.NewVBox(
		winContainer,
		layout.NewSpacer(),
		winHelp,
	)
}
