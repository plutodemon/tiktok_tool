package ui

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/google/gopacket/pcap"
	"tiktok_tool/config"
	"tiktok_tool/lkit"
)

type SettingsWindow struct {
	window fyne.Window
	// 网卡
	networkList     *widget.CheckGroup
	selectedDevices []string

	// 正则
	serverRegex    *widget.Entry
	streamKeyRegex *widget.Entry

	// 路径
	obsConfigPath     *widget.Entry
	liveCompanionPath *widget.Entry

	// 脚本路径
	pluginScriptPath *widget.Entry

	// 日志配置
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

	// 创建自动化插件脚本路径输入框
	w.pluginScriptPath = widget.NewEntry()
	w.pluginScriptPath.SetText(config.GetConfig().PluginScriptPath)
	w.pluginScriptPath.SetPlaceHolder("请选择自动化插件脚本路径 (auto.exe)")

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
	scriptTab := w.createScriptTab()
	otherTab := w.createOtherTab(&alreadyCheck)

	// 创建标签容器
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("正则设置", theme.DocumentIcon(), regexTab),
		container.NewTabItemWithIcon("网卡设置", theme.SearchIcon(), networkTab),
		container.NewTabItemWithIcon("日志设置", theme.ErrorIcon(), logTab),
		container.NewTabItemWithIcon("脚本设置", theme.ComputerIcon(), scriptTab),
		container.NewTabItemWithIcon("其他设置", theme.SettingsIcon(), otherTab),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	// 创建保存和取消按钮
	saveBtn := widget.NewButtonWithIcon("保存配置", theme.DocumentSaveIcon(), func() {
		w.saveSettings(alreadyCheck)
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButtonWithIcon("取消配置", theme.MailReplyIcon(), func() {
		w.window.Close()
	})

	// 创建按钮容器
	buttonContainer := container.NewHBox(
		saveBtn,
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

// createScriptTab 创建脚本设置标签页
func (w *SettingsWindow) createScriptTab() fyne.CanvasObject {
	// 创建插件脚本浏览按钮
	browsePluginScriptBtn := widget.NewButtonWithIcon("浏览", theme.FolderOpenIcon(), func() {
		w.browsePluginScriptConfig()
	})

	// 创建检测插件按钮
	detectBtn := widget.NewButtonWithIcon("检测插件", theme.SearchIcon(), func() {
		w.detectPlugin()
	})

	// 创建下载按钮
	downloadBtn := widget.NewButtonWithIcon("下载auto.exe", theme.DownloadIcon(), func() {
		w.downloadAutoExe()
	})
	downloadBtn.Importance = widget.MediumImportance

	// 创建清空插件按钮
	clearBtn := widget.NewButtonWithIcon("清空插件", theme.DeleteIcon(), func() {
		w.clearPlugin()
	})
	clearBtn.Importance = widget.DangerImportance

	// 创建插件脚本路径容器
	pluginScriptPathContainer := container.NewBorder(nil, nil, nil,
		container.NewHBox(browsePluginScriptBtn, detectBtn), w.pluginScriptPath)

	// 创建插件管理按钮容器
	pluginManageContainer := container.NewHBox(
		downloadBtn,
		clearBtn,
	)

	// 创建表单
	scriptForm := widget.NewForm(
		widget.NewFormItem("自动化插件脚本路径", pluginScriptPathContainer),
		widget.NewFormItem("插件管理", pluginManageContainer),
	)

	// 添加说明文本
	scriptHelp := widget.NewRichTextFromMarkdown("### 脚本设置说明\n\n" +
		"* **自动化插件脚本路径**：指定auto.exe文件路径\n\n" +
		"### 使用说明\n\n" +
		"1. 点击\"检测插件\"按钮可检测plugin目录下的插件文件\n" +
		"2. 点击\"清空插件\"按钮可删除plugin目录及所有内容\n" +
		"3. 配置脚本路径后，可以一键开播，实现自动化操作")

	// 创建容器
	return container.NewVBox(
		scriptForm,
		layout.NewSpacer(),
		scriptHelp,
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
func (w *SettingsWindow) createOtherTab(alreadyCheck *[]string) fyne.CanvasObject {
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

	// 创建恢复默认配置按钮
	resetBtn := widget.NewButtonWithIcon("恢复默认配置", theme.HistoryIcon(), func() {
		w.resetToDefaults(alreadyCheck)
	})

	// 添加说明文本
	otherHelp := widget.NewRichTextFromMarkdown("### 路径配置说明\n\n" +
		"* **OBS配置文件路径**：OBS Studio的配置文件路径，用于导入OBS推流设置\n" +
		"* **直播伴侣启动路径**：抖音直播伴侣的可执行文件路径，用于快速启动直播伴侣")

	// 创建容器
	return container.NewVBox(
		otherForm,
		layout.NewSpacer(),
		resetBtn,
		widget.NewSeparator(),
		otherHelp,
	)
}

// resetToDefaults 重置为默认设置
func (w *SettingsWindow) resetToDefaults(alreadyCheck *[]string) {
	w.NewConfirmDialog("确认", "确定要恢复默认配置吗？", func(ok bool) {
		if ok {
			// 检查并删除配置文件
			configPath := "config/tiktok_tool_cfg.toml"
			if _, err := os.Stat(configPath); err == nil {
				// 配置文件在 config 目录中
				if err := os.Remove(configPath); err != nil {
					w.NewErrorDialog(fmt.Errorf("删除配置文件失败: %v", err))
					return
				}
			} else {
				// 检查当前目录
				configPath = "tiktok_tool_cfg.toml"
				if _, err := os.Stat(configPath); err == nil {
					if err := os.Remove(configPath); err != nil {
						w.NewErrorDialog(fmt.Errorf("删除配置文件失败: %v", err))
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
			w.pluginScriptPath.SetText(config.DefaultConfig.PluginScriptPath)
			// 恢复日志配置默认值
			if config.DefaultConfig.LogConfig != nil {
				w.logToFile.SetChecked(config.DefaultConfig.LogConfig.File)
				w.logLevel.SetSelected(config.DefaultConfig.LogConfig.Level)
			}
			*alreadyCheck = nil // 清空已选网卡

			// 更新当前设置为默认设置
			config.SetConfig(config.DefaultConfig)

			w.NewInfoDialog("成功", "已恢复默认配置")
		}
	})
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
		PluginScriptPath:  strings.TrimSpace(w.pluginScriptPath.Text),
		LogConfig:         &updatedLogConfig,
	}

	// 保存设置
	if err := config.SaveSettings(newSettings); err != nil {
		w.NewErrorDialog(err)
		return
	}

	// 更新当前设置
	config.SetConfig(newSettings)

	w.NewInfoDialog("保存成功", "设置已保存，请重启软件以应用更改")
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
	fileDialog.SetConfirmText("确定选择")
	fileDialog.SetDismissText("取消选择")

	for {
		if path := config.GetConfig().OBSConfigPath; path != "" {
			pathDir := lkit.GetPathDir(path)
			if _, err := os.Stat(pathDir); err == nil {
				if uri := storage.NewFileURI(pathDir); uri != nil {
					if lister, err := storage.ListerForURI(uri); err == nil {
						fileDialog.SetLocation(lister)
						break
					}
				}
			}
		}
		defaultPath := filepath.Join(os.Getenv("APPDATA"), "obs-studio", "basic", "profiles")
		if _, err := os.Stat(defaultPath); err == nil {
			if uri := storage.NewFileURI(defaultPath); uri != nil {
				if lister, err := storage.ListerForURI(uri); err == nil {
					fileDialog.SetLocation(lister)
					break
				}
			}
		}
		break
	}

	fileDialog.Show()
}

// autoDetectOBSConfig 自动检测OBS配置文件路径
func (w *SettingsWindow) autoDetectOBSConfig() {
	detectedPath := config.GetDefaultOBSConfigPath()
	if detectedPath == "" {
		w.NewInfoDialog("检测结果", "未找到OBS配置文件，请确保OBS Studio已安装并至少运行过一次")
		return
	}

	w.obsConfigPath.SetText(detectedPath)
	w.NewInfoDialog("检测成功", fmt.Sprintf("已自动检测到OBS配置文件：\n%s", detectedPath))
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
	fileDialog.SetConfirmText("确定选择")
	fileDialog.SetDismissText("取消选择")

	if path := config.GetConfig().LiveCompanionPath; path != "" {
		pathDir := lkit.GetPathDir(path)
		if _, err := os.Stat(pathDir); err == nil {
			if uri := storage.NewFileURI(pathDir); uri != nil {
				if lister, err := storage.ListerForURI(uri); err == nil {
					fileDialog.SetLocation(lister)
				}
			}
		}
	}

	fileDialog.Show()
}

// detectPlugin 检测插件目录下的文件
func (w *SettingsWindow) detectPlugin() {
	pluginDir := "plugin"

	// 检查目录是否存在
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		w.NewInfoDialog("检测结果", "plugin目录不存在")
		return
	}

	// 读取目录内容
	files, err := os.ReadDir(pluginDir)
	if err != nil {
		w.NewErrorDialog(fmt.Errorf("读取plugin目录失败: %v", err))
		return
	}

	// 构建检测结果
	var result strings.Builder
	result.WriteString(fmt.Sprintf("plugin目录检测结果:\n\n"))

	if len(files) == 0 {
		result.WriteString("目录为空")
	} else {
		result.WriteString(fmt.Sprintf("共找到 %d 个文件/文件夹:\n\n", len(files)))
		for _, file := range files {
			if file.IsDir() {
				result.WriteString(fmt.Sprintf("📁 %s (文件夹)\n", file.Name()))
			} else {
				result.WriteString(fmt.Sprintf("📄 %s\n", file.Name()))
			}
			if strings.EqualFold(file.Name(), "auto.exe") {
				w.pluginScriptPath.SetText(filepath.Join(pluginDir, file.Name()))
			}
		}
	}

	w.NewInfoDialog("检测结果", result.String())
}

// clearPlugin 清空插件目录
func (w *SettingsWindow) clearPlugin() {
	pluginDir := "plugin"

	w.NewConfirmDialog(
		"确认清空",
		"确定要删除plugin目录及所有内容吗？\n此操作不可恢复！",
		func(confirmed bool) {
			if !confirmed {
				return
			}

			// 检查目录是否存在
			if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
				w.NewInfoDialog("清空结果", "plugin目录不存在，无需清空")
				return
			}

			// 删除目录及所有内容
			if err := os.RemoveAll(pluginDir); err != nil {
				w.NewErrorDialog(fmt.Errorf("删除目录失败: %v", err))
				return
			}

			// 清空插件脚本路径
			w.pluginScriptPath.SetText("")

			w.NewInfoDialog("清空结果", "plugin目录已成功删除")
		},
	)
}

// downloadAutoExe 下载auto.exe文件
func (w *SettingsWindow) downloadAutoExe() {
	// 创建进度对话框
	progressLabel := widget.NewLabel("准备下载...")
	progressBar := widget.NewProgressBar()
	progressDialog := w.NewCustomDialog("下载auto.exe", "取消", container.NewVBox(
		progressLabel,
		progressBar,
	))

	// 在后台执行下载
	lkit.SafeGo(func() {
		// 确保plugin目录存在
		pluginDir := "plugin"
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			progressDialog.Hide()
			w.NewErrorDialog(fmt.Errorf("创建plugin目录失败: %v", err))
			return
		}

		// 下载文件
		downloadURL := "https://github.com/plutodemon/py_win_auto/releases/download/v0.1.0/auto.exe"
		filePath := filepath.Join(pluginDir, "auto.exe")

		progressLabel.SetText("正在下载...")
		resp, err := http.Get(downloadURL)
		if err != nil {
			progressDialog.Hide()
			w.NewErrorDialog(fmt.Errorf("下载失败: %v", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			progressDialog.Hide()
			w.NewErrorDialog(fmt.Errorf("下载失败: HTTP %d", resp.StatusCode))
			return
		}

		// 创建目标文件
		file, err := os.Create(filePath)
		if err != nil {
			progressDialog.Hide()
			w.NewErrorDialog(fmt.Errorf("创建文件失败: %v", err))
			return
		}
		defer file.Close()

		// 复制文件内容
		progressLabel.SetText("正在保存文件...")
		progressBar.SetValue(0.7)
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			progressDialog.Hide()
			w.NewErrorDialog(fmt.Errorf("保存文件失败: %v", err))
			return
		}

		// 下载完成
		progressBar.SetValue(1.0)
		progressLabel.SetText("下载完成")
		time.Sleep(500 * time.Millisecond)
		progressDialog.Hide()

		// 自动设置路径
		w.pluginScriptPath.SetText(filePath)

		// 显示成功消息
		w.NewInfoDialog("下载成功", fmt.Sprintf("auto.exe已下载到:\n%s\n\n路径已自动设置到脚本路径中。", filePath))
	})
}

// browsePluginScriptConfig 浏览选择自动化插件脚本文件
func (w *SettingsWindow) browsePluginScriptConfig() {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if reader == nil {
			return
		}
		defer reader.Close()

		filePath := reader.URI().Path()
		w.pluginScriptPath.SetText(filePath)
	}, w.window)

	// 设置文件过滤器，支持多种脚本文件格式
	filter := storage.NewExtensionFileFilter([]string{".exe"})
	fileDialog.SetFilter(filter)
	fileDialog.SetConfirmText("确定选择")
	fileDialog.SetDismissText("取消选择")

	for {
		if path := config.GetConfig().PluginScriptPath; path != "" {
			pathDir := lkit.GetPathDir(path)
			if _, err := os.Stat(pathDir); err == nil {
				if uri := storage.NewFileURI(pathDir); uri != nil {
					if lister, err := storage.ListerForURI(uri); err == nil {
						fileDialog.SetLocation(lister)
						break
					}
				}
			}
		}
		// 设置默认路径为当前目录下的plugin文件夹
		currentDir, err := os.Getwd()
		if err != nil {
			break
		}
		pluginDir := filepath.Join(currentDir, "plugin")
		if _, err := os.Stat(pluginDir); err == nil {
			if uri := storage.NewFileURI(pluginDir); uri != nil {
				if lister, err := storage.ListerForURI(uri); err == nil {
					fileDialog.SetLocation(lister)
					break
				}
			}
		}
		// 如果plugin目录不存在，使用当前目录作为默认路径
		if uri := storage.NewFileURI(currentDir); uri != nil {
			if lister, err := storage.ListerForURI(uri); err == nil {
				fileDialog.SetLocation(lister)
				break
			}
		}
		break
	}

	fileDialog.Show()
}

// autoDetectLiveCompanionConfig 自动检测直播伴侣启动路径
func (w *SettingsWindow) autoDetectLiveCompanionConfig() {
	progressDialog := w.NewCustomWithoutButtons(
		"搜索中",
		container.NewCenter(widget.NewLabel("正在搜索直播伴侣安装路径，请稍候...")),
	)

	resultChan := make(chan string)

	lkit.SafeGo(func() {
		resultChan <- config.FindLiveCompanionPath()
	})

	select {
	case result := <-resultChan:
		progressDialog.Hide()
		if result == "" {
			w.NewInfoDialog("检测结果", "未找到直播伴侣安装路径。\n\n可能的原因：\n1. 抖音直播伴侣未安装\n2. 安装在非标准位置\n3. 文件名与预期不符\n\n请尝试手动浏览选择。")
			return
		}
		w.liveCompanionPath.SetText(result)
		w.NewInfoDialog("检测成功", fmt.Sprintf("已自动检测到直播伴侣启动路径：\n%s", result))

	case <-time.After(15 * time.Second):
		progressDialog.Hide()
		w.NewInfoDialog("检测结果", "未找到直播伴侣安装路径。\n\n可能的原因：\n1. 抖音直播伴侣未安装\n2. 安装在非标准位置\n3. 文件名与预期不符\n\n请尝试手动浏览选择。")
	}
}
