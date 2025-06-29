package ui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"os"
	"os/exec"
	"strings"
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
	w.settingBtn = widget.NewButtonWithIcon("设置", theme.SettingsIcon(), func() {
		ShowSettingsWindow(w.app)
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

		if config.IsDebug {
			serverRegexCompile := config.CurrentSettings.ServerRegex
			if len(serverRegexCompile) == 0 {
				serverRegexCompile = config.DefaultSettings.ServerRegex
			}
			fmt.Println("服务器地址正则表达式: ", serverRegexCompile)

			keyRegex := config.DefaultSettings.StreamKeyRegex
			if len(config.CurrentSettings.StreamKeyRegex) > 0 && config.CurrentSettings.StreamKeyRegex != keyRegex {
				keyRegex = config.CurrentSettings.StreamKeyRegex
			}
			fmt.Println("推流码正则表达式: ", keyRegex)
		}

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
			func() {
				w.status.SetText("已停止抓包")
				w.captureBtn.SetText("开始抓包")
				w.captureBtn.Importance = widget.HighImportance
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

// handleImportOBS 处理导入OBS配置
func (w *MainWindow) handleImportOBS() {
	// 检查是否有推流信息
	serverAddr := strings.TrimSpace(w.serverAddr.Text)
	streamKey := strings.TrimSpace(w.streamKey.Text)

	if serverAddr == "" || streamKey == "" {
		dialog.ShowInformation("提示", "请先抓取到推流服务器地址和推流码", w.window)
		return
	}

	// 检查OBS配置路径是否设置
	obsConfigPath := strings.TrimSpace(config.CurrentSettings.OBSConfigPath)
	if obsConfigPath == "" {
		dialog.ShowInformation("提示", "请先在设置中配置OBS配置文件路径", w.window)
		return
	}

	// 确认对话框
	dialog.ShowConfirm("确认导入",
		"将要导入配置到以下OBS配置文件中：\n"+obsConfigPath,
		func(confirmed bool) {
			if !confirmed {
				return
			}

			// 写入OBS配置
			err := config.WriteOBSConfig(obsConfigPath, serverAddr, streamKey)
			if err != nil {
				dialog.ShowError(fmt.Errorf("导入OBS配置失败：%v", err), w.window)
				return
			}

			w.status.SetText("OBS配置导入成功")
			dialog.ShowInformation("成功", "推流配置已成功导入到OBS！", w.window)
		}, w.window)
}
