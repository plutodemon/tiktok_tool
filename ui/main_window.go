package ui

import (
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"tiktok_tool/capture"
)

type MainWindow struct {
	window     fyne.Window
	app        fyne.App
	status     *widget.Label
	serverAddr *widget.Entry
	streamKey  *widget.Entry
	captureBtn *widget.Button
}

func (w *MainWindow) resetCaptureBtn() {
	w.captureBtn.SetText("开始抓包")
	w.captureBtn.Importance = widget.HighImportance
	w.captureBtn.SetIcon(theme.MediaPlayIcon())
}

//go:embed img/tiktok.png
var iconBytes []byte

func NewMainWindow() {
	myApp := app.NewWithID("com.lemon.tiktok_tool")
	myApp.SetIcon(&fyne.StaticResource{
		StaticName:    "icon",
		StaticContent: iconBytes,
	})
	window := myApp.NewWindow("抖音直播推流配置抓取")
	window.Resize(fyne.NewSize(600, 180))
	window.SetFixedSize(true)
	window.SetMaster()
	window.CenterOnScreen()

	if !capture.CheckNpcapInstalled() {
		ShowInstallDialog(window)
		window.ShowAndRun()
		return
	}

	w := &MainWindow{
		window:     window,
		app:        myApp,
		status:     widget.NewLabel("等待开始抓包..."),
		serverAddr: widget.NewEntry(),
		streamKey:  widget.NewEntry(),
	}

	w.setupUI()
	window.ShowAndRun()
	return
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
		if w.serverAddr.Text == "" {
			w.status.SetText("服务器地址为空")
			return
		}
		w.window.Clipboard().SetContent(w.serverAddr.Text)
		w.status.SetText("已复制服务器地址")
	})

	copyStreamBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		if w.streamKey.Text == "" {
			w.status.SetText("推流码为空")
			return
		}
		w.window.Clipboard().SetContent(w.streamKey.Text)
		w.status.SetText("已复制推流码")
	})

	// 启动直播伴侣按钮
	liveBtn := widget.NewButtonWithIcon("启动直播伴侣", theme.MediaPlayIcon(), w.handleStartLiveCompanion)

	// 导入OBS配置按钮
	importOBSBtn := widget.NewButtonWithIcon("导入OBS", theme.DocumentSaveIcon(), w.handleImportOBS)

	// 抓包按钮
	w.captureBtn = widget.NewButtonWithIcon("开始抓包", theme.MediaPlayIcon(), w.handleCapture)
	w.captureBtn.Importance = widget.HighImportance

	serverContainer := container.NewBorder(nil, nil, nil, copyServerBtn, w.serverAddr)
	streamContainer := container.NewBorder(nil, nil, nil, copyStreamBtn, w.streamKey)
	actionContainer := container.New(layout.NewGridLayout(3), liveBtn, w.captureBtn, importOBSBtn)

	mainForm := widget.NewForm(
		widget.NewFormItem("服务器地址", serverContainer),
		widget.NewFormItem("推流码", streamContainer),
	)

	// 重启按钮
	restartBtn := widget.NewButtonWithIcon("重启", theme.ViewRefreshIcon(), w.handleRestart)
	restartBtn.Importance = widget.LowImportance

	// 创建使用说明按钮
	helpBtn := widget.NewButtonWithIcon("帮助", theme.HelpIcon(), func() {
		ShowHelpDialog(w.window)
	})
	helpBtn.Importance = widget.LowImportance

	// 创建设置按钮
	settingBtn := widget.NewButtonWithIcon("设置", theme.SettingsIcon(), func() {
		ShowSettingsWindow(w.app, func() {
			w.window.Show()
		})
		w.window.Hide()
	})
	settingBtn.Importance = widget.LowImportance

	statusContainer := container.NewHBox(
		widget.NewIcon(theme.InfoIcon()),
		w.status,
		layout.NewSpacer(),
		restartBtn,
		helpBtn,
		settingBtn,
	)

	content := container.NewVBox(
		mainForm,
		actionContainer,
		layout.NewSpacer(),
		widget.NewSeparator(),
		statusContainer,
	)

	w.window.SetContent(container.NewPadded(content))
}
