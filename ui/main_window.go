package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"os"
	"os/exec"
	"tiktok_tool/capture"
	"tiktok_tool/config"
)

type MainWindow struct {
	window     fyne.Window
	app        fyne.App
	status     *widget.Label
	serverAddr *widget.Entry
	streamKey  *widget.Entry
	captureBtn *widget.Button
	settingBtn *widget.Button
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

	// 抓包按钮
	w.captureBtn = widget.NewButton("开始抓包", nil)
	w.captureBtn.Importance = widget.HighImportance
	w.captureBtn.Resize(w.captureBtn.MinSize())

	// 重启按钮
	restartBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), w.handleRestart)
	restartBtn.Importance = widget.WarningImportance
	restartBtn.Resize(restartBtn.MinSize())

	// 设置抓包按钮回调
	w.captureBtn.OnTapped = w.handleCapture

	// 创建固定宽度的标签
	serverLabel := widget.NewLabel("服务器地址:")
	streamLabel := widget.NewLabel("推 流 码 :") // 使用空格补齐，使其与"服务器地址"等宽

	// 创建分组：推流配置
	configTitle := container.NewBorder(
		nil, nil,
		widget.NewRichTextFromMarkdown("## 推流配置"),
		container.NewHBox(
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
	helpBtn := widget.NewButtonWithIcon("", theme.HelpIcon(), func() {
		ShowHelpDialog(w.window)
	})

	// 创建设置按钮
	w.settingBtn = widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		ShowSettingsWindow(w.app)
	})

	// 创建状态栏
	statusBar := container.NewHBox(
		widget.NewIcon(theme.InfoIcon()),
		w.status,
		helpBtn,
		w.settingBtn,
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

func (w *MainWindow) handleCapture() {
	if !config.IsCapturing {
		// 开始抓包
		config.IsCapturing = true
		config.StopCapture = make(chan struct{})

		// 更改按钮样式为停止状态
		w.captureBtn.SetText("停止抓包")
		w.captureBtn.Importance = widget.DangerImportance

		// 清空数据
		w.serverAddr.SetText("")
		w.streamKey.SetText("")

		w.status.SetText("正在抓包...")
		capture.StartCapture(
			func(server string) {
				w.serverAddr.SetText(server)
				w.status.SetText("已找到推流服务器地址")
			},
			func(key string) {
				w.streamKey.SetText(key)
				w.status.SetText("已找到推流码")
			},
			func(err error) {
				w.status.SetText("错误: " + err.Error())
			},
		)
	} else {
		// 停止抓包
		capture.StopCapturing()
		w.status.SetText("已停止抓包")

		// 更改按钮样式为开始状态
		w.captureBtn.SetText("开始抓包")
		w.captureBtn.Importance = widget.HighImportance
	}
}

func (w *MainWindow) handleRestart() {
	ShowRestartConfirmDialog(w.window, func() {
		if config.IsCapturing {
			capture.StopCapturing()
		}

		exe, err := os.Executable()
		if err != nil {
			dialog.ShowError(err, w.window)
			return
		}

		cmd := exec.Command(exe)
		err = cmd.Start()
		if err != nil {
			dialog.ShowError(err, w.window)
			return
		}

		w.app.Quit()
	})
}
