package ui

import (
	_ "embed"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"tiktok_tool/capture"
	"tiktok_tool/config"
	"tiktok_tool/lkit"
)

var (
	//go:embed img/tiktok.png
	tikTokIcon         []byte
	TikTokIconResource = fyne.NewStaticResource("tiktokIcon", tikTokIcon)

	//go:embed img/tiktok_dis.png
	tikTokIconDis         []byte
	TikTokIconResourceDis = fyne.NewStaticResource("tiktokIconDis", tikTokIconDis)

	//go:embed img/live.png
	LiveIcon         []byte
	LiveIconResource = fyne.NewStaticResource("liveIcon", LiveIcon)

	//go:embed img/OBS.png
	OBSIcon         []byte
	OBSIconResource = fyne.NewStaticResource("OBSIcon", OBSIcon)

	//go:embed img/OBS_dis.png
	OBSIconDis         []byte
	OBSIconResourceDis = fyne.NewStaticResource("OBSIconDis", OBSIconDis)

	//go:embed img/font.ttf
	resourceTtf  []byte
	resourceFont = fyne.NewStaticResource("font.ttf", resourceTtf)
)

type MainWindow struct {
	window     fyne.Window
	app        fyne.App
	serverAddr *widget.Entry
	streamKey  *widget.Entry

	captureBtn   *widget.Button
	importOBSBtn *widget.Button
	liveBtn      *widget.Button
	obsBtn       *widget.Button
	autoBtn      *widget.Button

	status     *widget.Label
	restartBtn *widget.Button
	settingBtn *widget.Button
}

type ChineseTheme struct{}

func (ChineseTheme) Font(fyne.TextStyle) fyne.Resource {
	return resourceFont
}
func (ChineseTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(n, v)
}
func (ChineseTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}
func (ChineseTheme) Size(n fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(n)
}

func NewMainWindow() {
	myApp := app.NewWithID("com.lemon.tiktok_tool")
	myApp.SetIcon(TikTokIconResource)
	myApp.Settings().SetTheme(&ChineseTheme{})
	window := myApp.NewWindow("抖音直播推流配置抓取")
	window.Resize(fyne.NewSize(600, 280))
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

	w.addSystemTray()
	window.SetCloseIntercept(w.handleWindowClose)
	w.setupUI()
	window.ShowAndRun()
}

// 系统托盘
func (w *MainWindow) addSystemTray() {
	desk, ok := w.app.(desktop.App)
	if !ok {
		return
	}

	menuItem1 := fyne.NewMenuItem("显示主窗口", func() {
		w.window.Show()
	})
	menuItem2 := fyne.NewMenuItem("启动直播伴侣", w.handleStartLiveCompanion)
	menuItem2.Disabled = !lkit.IsAdmin
	menuItem2.Icon = LiveIconResource

	m := fyne.NewMenu("tiktok_tool",
		menuItem1,
		fyne.NewMenuItemSeparator(),
		menuItem2,
	)
	desk.SetSystemTrayMenu(m)
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

	cfg := config.GetConfig()

	// 抓包按钮
	w.captureBtn = widget.NewButtonWithIcon("开始抓包", theme.MediaPlayIcon(), w.handleCapture)
	w.captureBtn.Importance = widget.HighImportance

	// 导入OBS配置按钮
	w.importOBSBtn = widget.NewButtonWithIcon("导入OBS", theme.DocumentSaveIcon(), w.handleImportOBS)
	if cfg.OBSConfigPath == "" {
		w.importOBSBtn.Disable()
	}

	// 启动直播伴侣按钮
	w.liveBtn = widget.NewButtonWithIcon("启动直播伴侣", LiveIconResource, w.handleStartLiveCompanion)
	if cfg.LiveCompanionPath == "" {
		w.liveBtn.Disable()
	}

	// 启动OBS
	w.obsBtn = widget.NewButtonWithIcon("启动OBS", OBSIconResource, w.handleStartOBS)
	if cfg.OBSLaunchPath == "" {
		w.obsBtn.Disable()
		w.obsBtn.SetIcon(OBSIconResourceDis)
	}

	// 自动推流按钮
	w.autoBtn = widget.NewButtonWithIcon("一键开播", TikTokIconResourceDis, nil)
	if lkit.IsAdmin && cfg.OBSConfigPath != "" && cfg.LiveCompanionPath != "" && cfg.OBSLaunchPath != "" {
		w.autoBtn.SetIcon(TikTokIconResource)
		w.autoBtn.OnTapped = w.handleAutoStart
	} else {
		w.autoBtn.Disable()
	}

	serverContainer := container.NewBorder(nil, nil, nil, copyServerBtn, w.serverAddr)
	streamContainer := container.NewBorder(nil, nil, nil, copyStreamBtn, w.streamKey)

	mainForm := widget.NewForm(
		widget.NewFormItem("服务器地址", serverContainer),
		widget.NewFormItem("推流码", streamContainer),
	)

	// 重启按钮
	w.restartBtn = widget.NewButtonWithIcon("重启", theme.ViewRefreshIcon(), func() {
		ShowRestartConfirmDialog(w.window, w.restartApp)
	})
	w.restartBtn.Importance = widget.LowImportance

	// 创建使用说明按钮
	helpBtn := widget.NewButtonWithIcon("帮助", theme.HelpIcon(), func() {
		ShowHelpDialog(w.window)
	})
	helpBtn.Importance = widget.LowImportance

	// 创建设置按钮
	w.settingBtn = widget.NewButtonWithIcon("设置", theme.SettingsIcon(), w.settingWindow)
	w.settingBtn.Importance = widget.LowImportance

	// 创建权限状态标签
	permissionStatus := widget.NewLabel("User")
	if lkit.IsAdmin {
		permissionStatus.SetText("Admin")
		permissionStatus.Importance = widget.SuccessImportance
	}

	actionContainer := container.New(layout.NewGridLayout(2), w.captureBtn, w.importOBSBtn)
	openContainer := container.New(layout.NewGridLayout(2), w.liveBtn, w.obsBtn)
	autoContainer := container.New(layout.NewGridLayout(1), w.autoBtn)

	statusContainer := container.NewHBox(
		widget.NewIcon(theme.InfoIcon()),
		permissionStatus,
		w.status,
		layout.NewSpacer(),
		w.restartBtn,
		helpBtn,
		w.settingBtn,
	)

	content := container.NewVBox(
		container.NewPadded(mainForm),
		container.NewPadded(actionContainer),
		container.NewPadded(openContainer),
		container.NewPadded(autoContainer),
		layout.NewSpacer(),
		widget.NewSeparator(),
		statusContainer,
	)

	w.window.SetContent(container.NewPadded(content))
}

func (w *MainWindow) settingWindow() {
	w.window.Hide()
	ShowSettingsWindow(w.app, func() { w.window.Show() }, func(text string) {
		dialog := w.NewCustomDialog("保存成功", "重启",
			container.NewCenter(widget.NewLabel(text)),
		)
		dialog.SetOnClosed(w.restartApp)
	})
}
