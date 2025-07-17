package ui

import (
	"net/url"
	"os/exec"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// parseURL 解析URL字符串
func parseURL(urlStr string) *url.URL {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil
	}
	return parsedURL
}

// ShowInstallDialog 显示Npcap安装对话框
func ShowInstallDialog(window fyne.Window) {
	content := container.NewVBox(
		widget.NewLabel("需要安装 Npcap 才能使用抓包功能："),
		widget.NewLabel("1. 点击下面的按钮下载 Npcap, 选择\"Npcap 1.80 installer for Windows即可\""),
		widget.NewButton("下载 Npcap", func() {
			_url := "https://npcap.com/#download"
			if err := exec.Command("cmd", "/c", "start", _url).Start(); err != nil {
				dialog.ShowError(err, window)
			}
		}),
		widget.NewLabel("2. 安装完成后重启程序"),
		widget.NewLabel("注意：安装时请勾选\"Install Npcap in WinPcap API-compatible Mode\""),
	)

	customDialog := dialog.NewCustom("安装提示", "关闭", content, window)
	customDialog.Resize(fyne.NewSize(600, 500))
	customDialog.Show()

	customDialog.SetOnClosed(func() {
		window.Close()
	})
}

// ShowHelpDialog 显示帮助对话框
func ShowHelpDialog(window fyne.Window) {
	// 创建超链接
	repoLink := widget.NewHyperlink("🔗 访问作者的GitHub仓库", parseURL("https://github.com/plutodemon/tiktok_tool"))

	// 创建声明部分
	disclaimer := widget.NewRichTextFromMarkdown("## ⚠️ 重要声明\n\n" +
		"**本软件为开源软件，不会收取任何费用！**\n\n" +
		"**本软件仅供学习交流使用，请勿用于商业用途！**")
	disclaimer.Wrapping = fyne.TextWrapWord

	// 创建使用说明
	usageGuide := widget.NewRichTextFromMarkdown("## 📖 使用方法\n\n" +
		"1. 点击 **开始抓包** 按钮\n" +
		"2. 打开抖音直播伴侣，开始直播\n" +
		"3. 等待自动获取推流配置\n" +
		"4. 如需停止请点击 **停止抓包** 按钮")
	usageGuide.Wrapping = fyne.TextWrapWord

	// 创建滚动容器
	scroll := container.NewScroll(
		container.NewVBox(
			widget.NewSeparator(),
			container.NewCenter(repoLink),
			widget.NewSeparator(),
			disclaimer,
			widget.NewSeparator(),
			usageGuide,
		),
	)

	content := container.NewBorder(
		nil,
		nil,
		nil,
		nil,
		scroll,
	)

	helpDialog := dialog.NewCustom("使用说明", "关闭", content, window)
	helpDialog.Resize(fyne.NewSize(450, 350))
	helpDialog.Show()
}

// ShowRestartConfirmDialog 显示重启确认对话框
func ShowRestartConfirmDialog(window fyne.Window, onConfirm func()) {
	confirmDialog := dialog.NewConfirm(
		"确认重启",
		"确定要重启程序吗? ",
		func(ok bool) {
			if ok {
				onConfirm()
			}
		},
		window,
	)
	confirmDialog.SetDismissText("取消")
	confirmDialog.SetConfirmText("重启")
	confirmDialog.Resize(MainWindowDialogSize)
	confirmDialog.Show()
}

var SettingsWindowDialogSize = fyne.NewSize(450, 200)

func (w *SettingsWindow) NewErrorDialog(err error) {
	errorDialog := dialog.NewError(err, w.window)
	errorDialog.Resize(SettingsWindowDialogSize)
	errorDialog.Show()
}

func (w *SettingsWindow) NewInfoDialog(title, message string) *dialog.Dialog {
	infoDialog := dialog.NewInformation(title, message, w.window)
	infoDialog.Resize(SettingsWindowDialogSize)
	infoDialog.Show()
	return &infoDialog
}

func (w *SettingsWindow) NewCustomDialog(title, dismiss string, content fyne.CanvasObject) *dialog.CustomDialog {
	customDialog := dialog.NewCustom(title, dismiss, content, w.window)
	customDialog.Resize(SettingsWindowDialogSize)
	customDialog.Show()
	return customDialog
}

func (w *SettingsWindow) NewConfirmDialog(title, message string, callback func(bool)) *dialog.ConfirmDialog {
	confirmDialog := dialog.NewConfirm(title, message, callback, w.window)
	confirmDialog.Resize(SettingsWindowDialogSize)
	confirmDialog.Show()
	return confirmDialog
}

func (w *SettingsWindow) NewCustomWithoutButtons(title string, content fyne.CanvasObject) *dialog.CustomDialog {
	customDialog := dialog.NewCustomWithoutButtons(title, content, w.window)
	customDialog.Resize(SettingsWindowDialogSize)
	customDialog.Show()
	return customDialog
}

var MainWindowDialogSize = fyne.NewSize(450, 200)

func (w *MainWindow) NewErrorDialog(err error) {
	errorDialog := dialog.NewError(err, w.window)
	errorDialog.Resize(MainWindowDialogSize)
	errorDialog.Show()
}

func (w *MainWindow) NewInfoDialog(title, message string) {
	infoDialog := dialog.NewInformation(title, message, w.window)
	infoDialog.Resize(MainWindowDialogSize)
	infoDialog.Show()
}

func (w *MainWindow) NewCustomDialog(title, dismiss string, content fyne.CanvasObject) *dialog.CustomDialog {
	customDialog := dialog.NewCustom(title, dismiss, content, w.window)
	customDialog.Resize(MainWindowDialogSize)
	customDialog.Show()
	return customDialog
}

func (w *MainWindow) NewConfirmDialog(title, message string, callback func(bool)) *dialog.ConfirmDialog {
	confirmDialog := dialog.NewConfirm(title, message, callback, w.window)
	confirmDialog.Resize(MainWindowDialogSize)
	confirmDialog.Show()
	return confirmDialog
}

func (w *MainWindow) NewCustomWithoutButtons(title string, content fyne.CanvasObject) *dialog.CustomDialog {
	customDialog := dialog.NewCustomWithoutButtons(title, content, w.window)
	customDialog.Resize(MainWindowDialogSize)
	customDialog.Show()
	return customDialog
}
