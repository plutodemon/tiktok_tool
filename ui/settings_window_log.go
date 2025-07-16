package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// createLogTab 创建日志设置标签页
func (w *SettingsWindow) createLogTab() fyne.CanvasObject {
	// 创建日志配置容器
	logConfigContainer := container.NewVBox(
		w.logToFile,
		container.NewHBox(
			widget.NewLabel("日志等级:"),
			w.logLevel,
		),
	)

	// 添加日志等级说明
	logLevelHelp := widget.NewRichTextFromMarkdown("### 日志等级说明\n\n" +
		"* **debug**: 最详细的日志信息，包含调试信息\n" +
		"* **info**: 一般信息日志（默认等级）\n" +
		"* **warn**: 警告信息\n" +
		"* **error**: 仅记录错误信息\n" +
		"### 注意:日志配置更改后需重启软件")

	// 创建容器
	return container.NewVBox(
		logConfigContainer,
		layout.NewSpacer(),
		logLevelHelp,
	)
}
