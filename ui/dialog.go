package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"os/exec"
)

// ShowInstallDialog 显示Npcap安装对话框
func ShowInstallDialog(window fyne.Window) {
	content := container.NewVBox(
		widget.NewLabel("需要安装 Npcap 才能使用抓包功能："),
		widget.NewLabel("1. 点击下面的按钮下载 Npcap, 选择\"Npcap 1.80 installer for Windows即可\""),
		widget.NewButton("下载 Npcap", func() {
			url := "https://npcap.com/#download"
			if err := exec.Command("cmd", "/c", "start", url).Start(); err != nil {
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
	scroll := container.NewScroll(
		container.NewVBox(
			widget.NewLabel("!!!声明!!!"),
			widget.NewLabel("!!!本软件为开源软件,不会收取任何费用!!!"),
			widget.NewLabel("!!!本软件仅供学习交流使用,请勿用于商业用途!!!"),
			widget.NewLabel("使用方法："),
			widget.NewLabel("1. 点击\"开始抓包\"按钮"),
			widget.NewLabel("2. 打开抖音直播伴侣, 开始直播"),
			widget.NewLabel("3. 等待自动获取推流配置"),
			widget.NewLabel("4. 如需停止请点击\"停止抓包\"按钮"),
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
	helpDialog.Resize(fyne.NewSize(370, 220))
	helpDialog.Show()
}

// ShowRestartConfirmDialog 显示重启确认对话框
func ShowRestartConfirmDialog(window fyne.Window, onConfirm func()) {
	confirmDialog := dialog.NewCustomConfirm(
		"确认重启",
		"确定",
		"取消",
		widget.NewLabel("确定要重启程序吗? "),
		func(ok bool) {
			if ok {
				onConfirm()
			}
		},
		window,
	)
	confirmDialog.Show()
}
