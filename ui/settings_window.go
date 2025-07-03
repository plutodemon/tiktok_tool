package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"tiktok_tool/lkit"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/google/gopacket/pcap"
	"tiktok_tool/config"
)

type SettingsWindow struct {
	window            fyne.Window
	networkList       *widget.CheckGroup
	selectedDevices   []string
	serverRegex       *widget.Entry
	streamKeyRegex    *widget.Entry
	obsConfigPath     *widget.Entry
	liveCompanionPath *widget.Entry
	// 日志配置相关字段
	logToFile *widget.Check
	logLevel  *widget.Select
}

func ShowSettingsWindow(parent fyne.App, closeCallback func()) {
	// 创建设置窗口
	settingsWindow := parent.NewWindow("设置")
	settingsWindow.Resize(fyne.NewSize(600, 350))
	settingsWindow.SetFixedSize(true)
	settingsWindow.CenterOnScreen()
	settingsWindow.SetOnClosed(closeCallback)

	sw := &SettingsWindow{
		window:          settingsWindow,
		selectedDevices: config.GetConfig().NetworkInterfaces,
	}

	sw.setupUI()
	settingsWindow.Show()
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

	// 创建OBS配置路径输入框
	w.obsConfigPath = widget.NewEntry()
	w.obsConfigPath.SetText(config.GetConfig().OBSConfigPath)
	w.obsConfigPath.SetPlaceHolder("请选择OBS配置文件路径 (service.json)")

	// 创建直播伴侣路径输入框
	w.liveCompanionPath = widget.NewEntry()
	w.liveCompanionPath.SetText(config.GetConfig().LiveCompanionPath)
	w.liveCompanionPath.SetPlaceHolder("请选择直播伴侣启动路径 (直播伴侣 Launcher.exe)")

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

	// 创建标签页内容
	regexTab := w.createRegexTab()
	networkTab := w.createNetworkTab()
	logTab := w.createLogTab()
	otherTab := w.createOtherTab()

	// 创建标签容器
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("正则设置", theme.DocumentIcon(), regexTab),
		container.NewTabItemWithIcon("网卡设置", theme.SearchIcon(), networkTab),
		container.NewTabItemWithIcon("日志设置", theme.ErrorIcon(), logTab),
		container.NewTabItemWithIcon("其他设置", theme.SettingsIcon(), otherTab),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	// 创建保存和取消按钮
	saveBtn := widget.NewButtonWithIcon("保存配置", theme.DocumentSaveIcon(), func() {
		w.saveSettings(alreadyCheck)
	})
	saveBtn.Importance = widget.HighImportance

	// 创建恢复默认配置按钮
	resetBtn := widget.NewButtonWithIcon("恢复默认配置", theme.HistoryIcon(), func() {
		w.resetToDefaults(&alreadyCheck)
	})
	resetBtn.Importance = widget.WarningImportance

	cancelBtn := widget.NewButtonWithIcon("取消配置", theme.MailReplyIcon(), func() {
		w.window.Close()
	})

	// 创建按钮容器
	buttonContainer := container.NewHBox(
		saveBtn,
		resetBtn,
		layout.NewSpacer(),
		cancelBtn,
	)

	// 设置内容
	w.window.SetContent(container.NewBorder(nil, buttonContainer, nil, nil, tabs))
}

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

// createLogTab 创建日志设置标签页
func (w *SettingsWindow) createLogTab() fyne.CanvasObject {
	// 创建日志配置容器
	logConfigContainer := container.NewVBox(
		w.logToFile,
		container.NewHBox(
			widget.NewLabel("日志等级:"),
			w.logLevel,
		),
	)

	// 添加日志等级说明
	logLevelHelp := widget.NewRichTextFromMarkdown("### 日志等级说明\n\n" +
		"* **debug**: 最详细的日志信息，包含调试信息\n" +
		"* **info**: 一般信息日志（默认等级）\n" +
		"* **warn**: 警告信息\n" +
		"* **error**: 仅记录错误信息\n" +
		"### 注意:日志配置更改后需重启软件")

	// 创建容器
	return container.NewVBox(
		logConfigContainer,
		layout.NewSpacer(),
		logLevelHelp,
	)
}

// createNetworkTab 创建网卡设置标签页
func (w *SettingsWindow) createNetworkTab() fyne.CanvasObject {
	// 创建网卡列表容器
	networkScroll := container.NewScroll(w.networkList)
	networkScroll.SetMinSize(fyne.NewSize(500, 170))

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

// createOtherTab 创建其他设置标签页
func (w *SettingsWindow) createOtherTab() fyne.CanvasObject {
	// 创建浏览按钮
	browseBtn := widget.NewButtonWithIcon("浏览", theme.FolderOpenIcon(), func() {
		w.browseOBSConfig()
	})

	// 创建自动检测按钮
	autoDetectBtn := widget.NewButtonWithIcon("自动检测", theme.SearchIcon(), func() {
		w.autoDetectOBSConfig()
	})

	// 创建直播伴侣浏览按钮
	browseLiveCompanionBtn := widget.NewButtonWithIcon("浏览", theme.FolderOpenIcon(), func() {
		w.browseLiveCompanionConfig()
	})

	// 创建直播伴侣自动检测按钮
	autoDetectLiveCompanionBtn := widget.NewButtonWithIcon("自动检测", theme.SearchIcon(), func() {
		w.autoDetectLiveCompanionConfig()
	})

	// 创建OBS配置路径容器
	obsPathContainer := container.NewBorder(nil, nil, nil,
		container.NewHBox(browseBtn, autoDetectBtn), w.obsConfigPath)

	// 创建直播伴侣路径容器
	liveCompanionPathContainer := container.NewBorder(nil, nil, nil,
		container.NewHBox(browseLiveCompanionBtn, autoDetectLiveCompanionBtn), w.liveCompanionPath)

	// 创建表单
	otherForm := widget.NewForm(
		widget.NewFormItem("OBS配置文件路径", obsPathContainer),
		widget.NewFormItem("直播伴侣启动路径", liveCompanionPathContainer),
	)

	// 添加说明文本
	otherHelp := widget.NewRichTextFromMarkdown("### 路径配置说明\n\n" +
		"* **OBS配置文件路径**：OBS Studio的配置文件路径，用于导入OBS推流设置\n" +
		"* **直播伴侣启动路径**：抖音直播伴侣的可执行文件路径，用于快速启动直播伴侣")

	// 创建容器
	return container.NewVBox(
		otherForm,
		layout.NewSpacer(),
		otherHelp,
	)
}

// resetToDefaults 重置为默认设置
func (w *SettingsWindow) resetToDefaults(alreadyCheck *[]string) {
	dialog.ShowConfirm("确认", "确定要恢复默认配置吗？", func(ok bool) {
		if ok {
			// 检查并删除配置文件
			configPath := "config/tiktok_tool_cfg.toml"
			if _, err := os.Stat(configPath); err == nil {
				// 配置文件在 config 目录中
				if err := os.Remove(configPath); err != nil {
					dialog.ShowError(fmt.Errorf("删除配置文件失败: %v", err), w.window)
					return
				}
			} else {
				// 检查当前目录
				configPath = "tiktok_tool_cfg.toml"
				if _, err := os.Stat(configPath); err == nil {
					if err := os.Remove(configPath); err != nil {
						dialog.ShowError(fmt.Errorf("删除配置文件失败: %v", err), w.window)
						return
					}
				}
			}

			// 恢复默认配置
			w.networkList.SetSelected(nil) // 清空网卡选择
			w.serverRegex.SetText(config.DefaultConfig.ServerRegex)
			w.streamKeyRegex.SetText(config.DefaultConfig.StreamKeyRegex)
			w.obsConfigPath.SetText(config.DefaultConfig.OBSConfigPath)
			w.liveCompanionPath.SetText(config.DefaultConfig.LiveCompanionPath)
			// 恢复日志配置默认值
			if config.DefaultConfig.LogConfig != nil {
				w.logToFile.SetChecked(config.DefaultConfig.LogConfig.File)
				w.logLevel.SetSelected(config.DefaultConfig.LogConfig.Level)
			}
			*alreadyCheck = nil // 清空已选网卡

			// 更新当前设置为默认设置
			config.SetConfig(config.DefaultConfig)

			dialog.ShowInformation("成功", "已恢复默认配置", w.window)
		}
	}, w.window)
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
		OBSConfigPath:     strings.TrimSpace(w.obsConfigPath.Text),
		LiveCompanionPath: strings.TrimSpace(w.liveCompanionPath.Text),
		LogConfig:         &updatedLogConfig,
	}

	// 保存设置
	if err := config.SaveSettings(newSettings); err != nil {
		dialog.ShowError(err, w.window)
		return
	}

	// 更新当前设置
	config.SetConfig(newSettings)

	dialog.ShowInformation("成功", "设置已保存", w.window)
	w.window.Close()
}

// browseOBSConfig 浏览选择OBS配置文件
func (w *SettingsWindow) browseOBSConfig() {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if reader == nil {
			return
		}
		defer reader.Close()

		filePath := reader.URI().Path()
		w.obsConfigPath.SetText(filePath)
	}, w.window)

	// 设置文件过滤器
	filter := storage.NewExtensionFileFilter([]string{".json"})
	fileDialog.SetFilter(filter)
	fileDialog.Show()
}

// autoDetectOBSConfig 自动检测OBS配置文件路径
func (w *SettingsWindow) autoDetectOBSConfig() {
	detectedPath := config.GetDefaultOBSConfigPath()
	if detectedPath == "" {
		dialog.ShowInformation("检测结果", "未找到OBS配置文件，请确保OBS Studio已安装并至少运行过一次", w.window)
		return
	}

	w.obsConfigPath.SetText(detectedPath)
	dialog.ShowInformation("检测成功", fmt.Sprintf("已自动检测到OBS配置文件：\n%s", detectedPath), w.window)
}

// browseLiveCompanionConfig 浏览选择直播伴侣启动文件
func (w *SettingsWindow) browseLiveCompanionConfig() {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if reader == nil {
			return
		}
		defer reader.Close()

		filePath := reader.URI().Path()
		w.liveCompanionPath.SetText(filePath)
	}, w.window)

	// 设置文件过滤器
	filter := storage.NewExtensionFileFilter([]string{".exe"})
	fileDialog.SetFilter(filter)
	fileDialog.Show()
}

// autoDetectLiveCompanionConfig 自动检测直播伴侣启动路径
func (w *SettingsWindow) autoDetectLiveCompanionConfig() {
	progressDialog := dialog.NewCustomWithoutButtons("搜索中", widget.NewLabel("正在搜索直播伴侣安装路径，请稍候..."), w.window)
	progressDialog.Show()

	resultChan := make(chan string)

	lkit.SafeGo(func() {
		resultChan <- config.FindLiveCompanionPath()
	})

	select {
	case result := <-resultChan:
		progressDialog.Hide()
		if result == "" {
			dialog.ShowInformation("检测结果", "未找到直播伴侣安装路径。\n\n可能的原因：\n1. 抖音直播伴侣未安装\n2. 安装在非标准位置\n3. 文件名与预期不符\n\n请尝试手动浏览选择。", w.window)
			return
		}
		w.liveCompanionPath.SetText(result)
		dialog.ShowInformation("检测成功", fmt.Sprintf("已自动检测到直播伴侣启动路径：\n%s", result), w.window)

	case <-time.After(15 * time.Second):
		progressDialog.Hide()
		dialog.ShowInformation("检测结果", "未找到直播伴侣安装路径。\n\n可能的原因：\n1. 抖音直播伴侣未安装\n2. 安装在非标准位置\n3. 文件名与预期不符\n\n请尝试手动浏览选择。", w.window)
	}
}
