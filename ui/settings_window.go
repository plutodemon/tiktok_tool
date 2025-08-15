package ui

import (
	"fmt"
	"strconv"
	"strings"

	"tiktok_tool/lkit"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/google/gopacket/pcap"

	"tiktok_tool/config"
)

type NumericalEntry struct {
	widget.Entry
}

func NewNumericalEntry() *NumericalEntry {
	entry := &NumericalEntry{}
	entry.ExtendBaseWidget(entry)
	return entry
}

func (e *NumericalEntry) TypedRune(r rune) {
	if r >= '0' && r <= '9' {
		e.Entry.TypedRune(r)
	}
}

func (e *NumericalEntry) TypedShortcut(shortcut fyne.Shortcut) {
	paste, ok := shortcut.(*fyne.ShortcutPaste)
	if !ok {
		e.Entry.TypedShortcut(shortcut)
		return
	}

	content := paste.Clipboard.Content()
	if _, err := strconv.ParseFloat(content, 64); err == nil {
		e.Entry.TypedShortcut(shortcut)
	}
}

type SettingsWindow struct {
	window        fyne.Window
	closeCallback func()
	saveCallback  func(string)
	// 网卡
	networkList     *widget.CheckGroup
	selectedDevices []string

	// 正则
	serverRegex    *widget.Entry
	streamKeyRegex *widget.Entry

	// 路径
	obsLaunchPath     *widget.Entry
	obsConfigPath     *widget.Entry
	liveCompanionPath *widget.Entry

	// 脚本路径
	pluginScriptPath        *widget.Entry
	pluginScriptDownloadBtn *widget.Button
	pluginCheckInterval     *NumericalEntry
	pluginWaitAfterFound    *NumericalEntry
	pluginTimeout           *NumericalEntry

	// 日志配置
	logToFile *widget.Check
	logLevel  *widget.Select

	// 窗口行为配置
	minimizeOnClose   *widget.Check
	openLiveWhenStart *widget.Check
	obsWsIp           *widget.Entry
	obsWsPort         *widget.Entry
	obsWsPassword     *widget.Entry
}

func ShowSettingsWindow(parent fyne.App, closeCallback func(), saveCallback func(string)) {
	// 创建设置窗口
	settingsWindow := parent.NewWindow("设置")
	settingsWindow.Resize(fyne.NewSize(605, 430))
	settingsWindow.SetFixedSize(true)
	settingsWindow.CenterOnScreen()

	sw := &SettingsWindow{
		window:          settingsWindow,
		closeCallback:   closeCallback,
		saveCallback:    saveCallback,
		selectedDevices: config.GetConfig().BaseSettings.NetworkInterfaces,
	}

	settingsWindow.SetCloseIntercept(func() {
		sw.close()
	})

	sw.setupUI()
	settingsWindow.Show()
}

func (w *SettingsWindow) close() {
	w.window.Close()
	w.closeCallback()
}

// setupUI 设置用户界面
func (w *SettingsWindow) setupUI() {
	// 获取所有网卡
	devices, _ := pcap.FindAllDevs()

	names := make([]string, 0)
	for _, device := range devices {
		names = append(names, device.Description)
	}

	// 创建网卡列表
	alreadyCheck := make([]string, 0)
	w.networkList = widget.NewCheckGroup(names, func(check []string) {
		alreadyCheck = check
	})
	w.networkList.SetSelected(w.selectedDevices)

	// 创建正则表达式输入框
	cfg := config.GetConfig()
	w.serverRegex = widget.NewMultiLineEntry()
	w.serverRegex.SetText(cfg.BaseSettings.ServerRegex)
	w.serverRegex.Wrapping = fyne.TextWrapBreak
	w.serverRegex.Resize(fyne.NewSize(w.serverRegex.Size().Width, 80))

	w.streamKeyRegex = widget.NewMultiLineEntry()
	w.streamKeyRegex.SetText(cfg.BaseSettings.StreamKeyRegex)
	w.streamKeyRegex.Wrapping = fyne.TextWrapBreak
	w.streamKeyRegex.Resize(fyne.NewSize(w.streamKeyRegex.Size().Width, 80))

	// 创建OBS启动路径输入框
	w.obsLaunchPath = widget.NewEntry()
	w.obsLaunchPath.SetText(cfg.PathSettings.OBSLaunchPath)
	w.obsLaunchPath.SetPlaceHolder("请选择OBS启动路径 (obs64.exe)")
	w.obsLaunchPath.Disable()

	// 创建OBS配置路径输入框
	w.obsConfigPath = widget.NewEntry()
	w.obsConfigPath.SetText(cfg.PathSettings.OBSConfigPath)
	w.obsConfigPath.SetPlaceHolder("请选择OBS配置文件路径 (service.json)")
	w.obsConfigPath.Disable()

	// 创建直播伴侣路径输入框
	w.liveCompanionPath = widget.NewEntry()
	w.liveCompanionPath.SetText(cfg.PathSettings.LiveCompanionPath)
	w.liveCompanionPath.SetPlaceHolder("请选择直播伴侣启动路径 (直播伴侣 Launcher.exe)")
	w.liveCompanionPath.Disable()

	// 创建自动化插件脚本路径输入框
	w.pluginScriptPath = widget.NewEntry()
	w.pluginScriptPath.SetText(cfg.PathSettings.PluginScriptPath)
	w.pluginScriptPath.SetPlaceHolder("请选择自动化插件脚本路径 (auto.exe)")
	w.pluginScriptPath.Disable()

	// 创建插件相关配置控件
	w.pluginCheckInterval = NewNumericalEntry()
	w.pluginCheckInterval.SetText(lkit.AnyToStr(cfg.ScriptSettings.PluginCheckInterval))
	w.pluginCheckInterval.SetPlaceHolder("插件检查间隔 (秒)")

	w.pluginWaitAfterFound = NewNumericalEntry()
	w.pluginWaitAfterFound.SetText(lkit.AnyToStr(cfg.ScriptSettings.PluginWaitAfterFound))
	w.pluginWaitAfterFound.SetPlaceHolder("插件检测到后等待时间 (秒)")

	w.pluginTimeout = NewNumericalEntry()
	w.pluginTimeout.SetText(lkit.AnyToStr(cfg.ScriptSettings.PluginTimeout))
	w.pluginTimeout.SetPlaceHolder("插件超时时间 (秒)")

	// 创建下载按钮
	w.pluginScriptDownloadBtn = widget.NewButtonWithIcon("下载auto.exe", theme.DownloadIcon(), w.downloadAutoExe)

	// 创建日志配置控件
	w.logToFile = widget.NewCheck("输出到文件", nil)
	if cfg.LogConfig != nil {
		w.logToFile.SetChecked(cfg.LogConfig.File)
	}

	w.logLevel = widget.NewSelect([]string{"debug", "info", "warn", "error"}, nil)
	if cfg.LogConfig != nil {
		w.logLevel.SetSelected(cfg.LogConfig.Level)
	} else {
		w.logLevel.SetSelected("info")
	}

	// 创建窗口行为配置控件
	w.minimizeOnClose = widget.NewCheck("关闭窗口时最小化到托盘", nil)
	w.minimizeOnClose.SetChecked(cfg.BaseSettings.MinimizeOnClose)

	w.openLiveWhenStart = widget.NewCheck("启动时打开直播伴侣以及OBS", nil)
	w.openLiveWhenStart.SetChecked(cfg.BaseSettings.OpenLiveWhenStart)

	// 创建OBS WebSocket配置控件
	w.obsWsIp = widget.NewEntry()
	w.obsWsIp.SetText(cfg.BaseSettings.OBSWsIp)
	w.obsWsIp.SetPlaceHolder("IP(本机:127.0.0.1)")

	w.obsWsPort = widget.NewEntry()
	if cfg.BaseSettings.OBSWsPort != 0 {
		w.obsWsPort.SetText(lkit.AnyToStr(cfg.BaseSettings.OBSWsPort))
	}
	w.obsWsPort.SetPlaceHolder("端口(默认:4455)")

	w.obsWsPassword = widget.NewEntry()
	w.obsWsPassword.SetText(cfg.BaseSettings.OBSWsPassword)
	w.obsWsPassword.SetPlaceHolder("密码")

	// 创建标签页内容
	regexTab := w.createRegexTab()
	networkTab := w.createNetworkTab()
	logTab := w.createLogTab()
	scriptTab := w.createScriptTab()
	pathTab := w.createPathTab(&alreadyCheck)
	windowTab := w.createWindowTab()

	// 创建标签容器
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("正则设置", theme.DocumentIcon(), regexTab),
		container.NewTabItemWithIcon("网卡设置", theme.SearchIcon(), networkTab),
		container.NewTabItemWithIcon("日志设置", theme.ErrorIcon(), logTab),
		container.NewTabItemWithIcon("脚本设置", theme.ComputerIcon(), scriptTab),
		container.NewTabItemWithIcon("路径设置", theme.SettingsIcon(), pathTab),
		container.NewTabItemWithIcon("窗口行为", theme.WindowMaximizeIcon(), windowTab),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	// 创建保存和取消按钮
	saveBtn := widget.NewButtonWithIcon("保存配置", theme.DocumentSaveIcon(), func() {
		w.saveSettings(alreadyCheck)
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButtonWithIcon("取消配置", theme.MailReplyIcon(), func() {
		w.close()
	})

	// 创建按钮容器
	buttonContainer := container.New(
		layout.NewGridLayout(2),
		saveBtn,
		cancelBtn,
	)

	// 设置内容
	w.window.SetContent(container.NewBorder(nil, buttonContainer, nil, nil, tabs))
}

// saveSettings 保存设置并进行验证
func (w *SettingsWindow) saveSettings(checks []string) {
	// 验证插件设置
	if err := w.validatePluginSettings(); err != nil {
		w.NewErrorDialog(err)
		return
	}

	// 获取当前配置以保留其他日志设置
	currentConfig := config.GetConfig()
	logConfig := currentConfig.LogConfig
	if logConfig == nil {
		logConfig = config.DefaultConfig.LogConfig
	}

	// 更新日志配置
	updatedLogConfig := *logConfig
	updatedLogConfig.File = w.logToFile.Checked
	updatedLogConfig.Level = w.logLevel.Selected

	// 创建新的设置
	newSettings := &config.Config{
		BaseSettings: &config.BaseSettings{
			NetworkInterfaces: checks,
			ServerRegex:       strings.TrimSpace(w.serverRegex.Text),
			StreamKeyRegex:    strings.TrimSpace(w.streamKeyRegex.Text),
			MinimizeOnClose:   w.minimizeOnClose.Checked,
			OpenLiveWhenStart: w.openLiveWhenStart.Checked,
			OBSWsIp:           strings.TrimSpace(w.obsWsIp.Text),
			OBSWsPort:         lkit.Str2Int32(w.obsWsPort.Text),
			OBSWsPassword:     strings.TrimSpace(w.obsWsPassword.Text),
		},
		PathSettings: &config.PathSettings{
			OBSLaunchPath:     strings.TrimSpace(w.obsLaunchPath.Text),
			OBSConfigPath:     strings.TrimSpace(w.obsConfigPath.Text),
			LiveCompanionPath: strings.TrimSpace(w.liveCompanionPath.Text),
			PluginScriptPath:  strings.TrimSpace(w.pluginScriptPath.Text),
		},
		ScriptSettings: &config.ScriptSettings{
			PluginCheckInterval:  lkit.Str2Int32(w.pluginCheckInterval.Text),
			PluginWaitAfterFound: lkit.Str2Int32(w.pluginWaitAfterFound.Text),
			PluginTimeout:        lkit.Str2Int32(w.pluginTimeout.Text),
		},
		LogConfig: &updatedLogConfig,
	}

	// 保存设置
	if err := config.SaveSettings(newSettings); err != nil {
		w.NewErrorDialog(err)
		return
	}

	w.close()
	w.saveCallback("设置已保存，请重启软件以应用更改")
}

// validatePluginSettings 验证插件设置
func (w *SettingsWindow) validatePluginSettings() error {
	// 验证插件检查间隔
	checkIntervalText := strings.TrimSpace(w.pluginCheckInterval.Text)
	if checkIntervalText == "" {
		return fmt.Errorf("插件检查间隔不能为空")
	}
	checkInterval := lkit.Str2Int32(checkIntervalText)
	if checkInterval < 1 {
		return fmt.Errorf("插件检查间隔不可小于1秒")
	}

	// 验证等待时间
	waitTimeText := strings.TrimSpace(w.pluginWaitAfterFound.Text)
	if waitTimeText == "" {
		return fmt.Errorf("等待时间不可为空")
	}
	waitTime := lkit.Str2Int32(waitTimeText)
	if waitTime < 0 {
		return fmt.Errorf("等待时间不能为负数")
	}

	// 验证超时时间
	timeoutText := strings.TrimSpace(w.pluginTimeout.Text)
	if timeoutText == "" {
		return fmt.Errorf("超时时间不能为空")
	}
	timeout := lkit.Str2Int32(timeoutText)
	if timeout < 5 {
		return fmt.Errorf("超时时间不可小于5秒")
	}
	if timeout > 300 {
		return fmt.Errorf("超时时间不可大于300秒")
	}

	return nil
}
