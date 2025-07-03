package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type MainWindow struct {
	window     fyne.Window
	app        fyne.App
	status     *widget.Label
	serverAddr *widget.Entry
	streamKey  *widget.Entry
	liveBtn    *widget.Button
	captureBtn *widget.Button
	settingBtn *widget.Button
}

func (w *MainWindow) resetCaptureBtn() {
	w.captureBtn.SetText("开始抓包")
	w.captureBtn.Importance = widget.HighImportance
	w.captureBtn.SetIcon(theme.MediaPlayIcon())
}

func NewMainWindow(app fyne.App, window fyne.Window) *MainWindow {
	w := &MainWindow{
		window:     window,
		app:        app,
		status:     widget.NewLabel("等待开始抓包..."),
		serverAddr: widget.NewEntry(),
		streamKey:  widget.NewEntry(),
	}
	w.setupUI()
	return w
}

func (w *MainWindow) setupUI() {
	// 设置输入框
	w.serverAddr.SetPlaceHolder("服务器地址")
	w.serverAddr.Resize(fyne.NewSize(200, w.serverAddr.MinSize().Height))
	w.serverAddr.Disable()
	w.streamKey.SetPlaceHolder("推流码")
	w.streamKey.Resize(fyne.NewSize(200, w.streamKey.MinSize().Height))
	w.streamKey.Disable()

	// 复制按钮
	copyServerBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		w.window.Clipboard().SetContent(w.serverAddr.Text)
		w.status.SetText("已复制服务器地址")
	})

	copyStreamBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		w.window.Clipboard().SetContent(w.streamKey.Text)
		w.status.SetText("已复制推流码")
	})

	// 启动直播伴侣按钮
	w.liveBtn = widget.NewButtonWithIcon("启动直播伴侣", theme.MediaPlayIcon(), w.handleStartLiveCompanion)
	w.liveBtn.Importance = widget.SuccessImportance
	w.liveBtn.Resize(w.liveBtn.MinSize())

	// 抓包按钮
	w.captureBtn = widget.NewButtonWithIcon("开始抓包", theme.MediaPlayIcon(), w.handleCapture)
	w.captureBtn.Importance = widget.HighImportance
	w.captureBtn.Resize(w.captureBtn.MinSize())

	// 重启按钮
	restartBtn := widget.NewButtonWithIcon("重启", theme.ViewRefreshIcon(), w.handleRestart)
	restartBtn.Importance = widget.WarningImportance
	restartBtn.Resize(restartBtn.MinSize())

	// 创建固定宽度的标签
	serverLabel := widget.NewLabel("服务器地址:")
	streamLabel := widget.NewLabel("推 流 码 :")

	// 创建分组：推流配置
	configTitle := container.NewBorder(
		nil, nil,
		widget.NewRichTextFromMarkdown("## 推流配置"),
		container.NewHBox(
			w.liveBtn,
			w.captureBtn,
			restartBtn,
		),
	)

	configGroup := widget.NewCard("", "", container.NewVBox(
		configTitle,
		container.NewBorder(nil, nil, serverLabel, copyServerBtn,
			w.serverAddr,
		),
		widget.NewSeparator(),
		container.NewBorder(nil, nil, streamLabel, copyStreamBtn,
			w.streamKey,
		),
	))

	// 创建使用说明按钮
	helpBtn := widget.NewButtonWithIcon("帮助", theme.HelpIcon(), func() {
		ShowHelpDialog(w.window)
	})

	// 创建设置按钮
	w.settingBtn = widget.NewButtonWithIcon("设置", theme.SettingsIcon(), func() {
		ShowSettingsWindow(w.app, func() {
			w.window.Show()
		})
		w.window.Hide()
	})

	// 导入OBS配置按钮
	importOBSBtn := widget.NewButtonWithIcon("导入OBS", theme.DocumentSaveIcon(), w.handleImportOBS)
	importOBSBtn.Importance = widget.MediumImportance

	left := container.NewHBox(
		widget.NewIcon(theme.InfoIcon()),
		w.status,
	)
	right := container.NewHBox(
		helpBtn,
		w.settingBtn,
		importOBSBtn,
	)

	// 创建状态栏
	statusBar := container.NewBorder(
		nil,
		nil,
		left,
		right,
	)

	content := container.NewVBox(
		configGroup,
		widget.NewSeparator(),
		statusBar,
	)

	// 添加内边距
	paddedContent := container.NewPadded(content)

	w.window.SetContent(paddedContent)
	w.window.Resize(fyne.NewSize(600, 200))
	w.window.CenterOnScreen()
}
