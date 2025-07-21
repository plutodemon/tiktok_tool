package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/google/gopacket/pcap"

	"tiktok_tool/config"
)

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

	// 日志配置
	logToFile *widget.Check
	logLevel  *widget.Select

	// 窗口行为配置
	minimizeOnClose *widget.Check
}

func ShowSettingsWindow(parent fyne.App, closeCallback func(), saveCallback func(string)) {
	// 创建设置窗口
	settingsWindow := parent.NewWindow("设置")
	settingsWindow.Resize(fyne.NewSize(605, 350))
	settingsWindow.SetFixedSize(true)
	settingsWindow.CenterOnScreen()

	sw := &SettingsWindow{
		window:          settingsWindow,
		closeCallback:   closeCallback,
		saveCallback:    saveCallback,
		selectedDevices: config.GetConfig().NetworkInterfaces,
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
	w.serverRegex = widget.NewMultiLineEntry()
	w.serverRegex.SetText(config.GetConfig().ServerRegex)
	w.serverRegex.Wrapping = fyne.TextWrapBreak
	w.serverRegex.Resize(fyne.NewSize(w.serverRegex.Size().Width, 80))

	w.streamKeyRegex = widget.NewMultiLineEntry()
	w.streamKeyRegex.SetText(config.GetConfig().StreamKeyRegex)
	w.streamKeyRegex.Wrapping = fyne.TextWrapBreak
	w.streamKeyRegex.Resize(fyne.NewSize(w.streamKeyRegex.Size().Width, 80))

	// 创建OBS启动路径输入框
	w.obsLaunchPath = widget.NewEntry()
	w.obsLaunchPath.SetText(config.GetConfig().OBSLaunchPath)
	w.obsLaunchPath.SetPlaceHolder("请选择OBS启动路径 (obs64.exe)")
	w.obsLaunchPath.Disable()

	// 创建OBS配置路径输入框
	w.obsConfigPath = widget.NewEntry()
	w.obsConfigPath.SetText(config.GetConfig().OBSConfigPath)
	w.obsConfigPath.SetPlaceHolder("请选择OBS配置文件路径 (service.json)")
	w.obsConfigPath.Disable()

	// 创建直播伴侣路径输入框
	w.liveCompanionPath = widget.NewEntry()
	w.liveCompanionPath.SetText(config.GetConfig().LiveCompanionPath)
	w.liveCompanionPath.SetPlaceHolder("请选择直播伴侣启动路径 (直播伴侣 Launcher.exe)")
	w.liveCompanionPath.Disable()

	// 创建自动化插件脚本路径输入框
	w.pluginScriptPath = widget.NewEntry()
	w.pluginScriptPath.SetText(config.GetConfig().PluginScriptPath)
	w.pluginScriptPath.SetPlaceHolder("请选择自动化插件脚本路径 (auto.exe)")
	w.pluginScriptPath.Disable()

	// 创建下载按钮
	w.pluginScriptDownloadBtn = widget.NewButtonWithIcon("下载auto.exe", theme.DownloadIcon(), w.downloadAutoExe)

	// 创建日志配置控件
	currentConfig := config.GetConfig()
	w.logToFile = widget.NewCheck("输出到文件", nil)
	if currentConfig.LogConfig != nil {
		w.logToFile.SetChecked(currentConfig.LogConfig.File)
	}

	w.logLevel = widget.NewSelect([]string{"debug", "info", "warn", "error"}, nil)
	if currentConfig.LogConfig != nil {
		w.logLevel.SetSelected(currentConfig.LogConfig.Level)
	} else {
		w.logLevel.SetSelected("info")
	}

	// 创建窗口行为配置控件
	w.minimizeOnClose = widget.NewCheck("关闭窗口时最小化到托盘", nil)
	w.minimizeOnClose.SetChecked(currentConfig.MinimizeOnClose)

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

func (w *SettingsWindow) saveSettings(checks []string) {
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
	newSettings := config.Config{
		NetworkInterfaces: checks,
		ServerRegex:       strings.TrimSpace(w.serverRegex.Text),
		StreamKeyRegex:    strings.TrimSpace(w.streamKeyRegex.Text),
		OBSLaunchPath:     strings.TrimSpace(w.obsLaunchPath.Text),
		OBSConfigPath:     strings.TrimSpace(w.obsConfigPath.Text),
		LiveCompanionPath: strings.TrimSpace(w.liveCompanionPath.Text),
		PluginScriptPath:  strings.TrimSpace(w.pluginScriptPath.Text),
		MinimizeOnClose:   w.minimizeOnClose.Checked,
		LogConfig:         &updatedLogConfig,
	}

	// 保存设置
	if err := config.SaveSettings(newSettings); err != nil {
		w.NewErrorDialog(err)
		return
	}

	w.close()
	w.saveCallback("设置已保存，请重启软件以应用更改")
}
