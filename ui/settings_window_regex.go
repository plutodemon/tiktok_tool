package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// createRegexTab 创建正则设置标签页
func (w *SettingsWindow) createRegexTab() fyne.CanvasObject {
	// 创建正则表达式表单
	regexForm := widget.NewForm(
		widget.NewFormItem("服务器地址正则", w.serverRegex),
		widget.NewFormItem("推流码正则", w.streamKeyRegex),
	)

	// 添加说明文本
	regexHelp := widget.NewRichTextFromMarkdown("### 正则表达式说明\n\n" +
		"* **服务器地址正则**：用于匹配抓包数据中的推流服务器地址\n" +
		"* **推流码正则**：用于匹配抓包数据中的推流密钥\n\n" +
		"正则表达式需要包含一个捕获组，用于提取匹配的内容。")

	// 创建容器
	return container.NewVBox(
		regexForm,
		layout.NewSpacer(),
		regexHelp,
	)
}
