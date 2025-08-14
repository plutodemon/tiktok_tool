package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// createNetworkTab 创建网卡设置标签页
func (w *SettingsWindow) createNetworkTab() fyne.CanvasObject {
	// 创建网卡列表容器
	networkScroll := container.NewScroll(w.networkList)
	networkScroll.SetMinSize(fyne.NewSize(500, 250))

	// 添加说明文本
	networkHelp := widget.NewRichTextFromMarkdown("### 网卡选择说明\n\n" +
		"选择需要监听的网卡，抓包功能将监听所选网卡的网络流量。\n\n" +
		"如果不确定使用哪个网卡，可以选择多个网卡同时监听。")

	// 创建容器
	return container.NewVBox(
		networkScroll,
		layout.NewSpacer(),
		networkHelp,
	)
}
