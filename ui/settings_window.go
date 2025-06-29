package ui

import (
	"fmt"
	"fyne.io/fyne/v2/storage"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/google/gopacket/pcap"
	"tiktok_tool/config"
)

type SettingsWindow struct {
	window          fyne.Window
	networkList     *widget.CheckGroup
	selectedDevices []string
	serverRegex     *widget.Entry
	streamKeyRegex  *widget.Entry
	obsConfigPath   *widget.Entry
}

func ShowSettingsWindow(parent fyne.App) {
	// 创建设置窗口
	settingsWindow := parent.NewWindow("设置")
	settingsWindow.Resize(fyne.NewSize(600, 600))
	settingsWindow.SetFixedSize(true)
	settingsWindow.CenterOnScreen()

	sw := &SettingsWindow{
		window:          settingsWindow,
		selectedDevices: config.CurrentSettings.NetworkInterfaces,
	}

	sw.setupUI()
	settingsWindow.Show()
}

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
	w.serverRegex.SetText(config.CurrentSettings.ServerRegex)
	w.serverRegex.Wrapping = fyne.TextWrapBreak
	w.serverRegex.Resize(fyne.NewSize(w.serverRegex.Size().Width, 50))

	w.streamKeyRegex = widget.NewMultiLineEntry()
	w.streamKeyRegex.SetText(config.CurrentSettings.StreamKeyRegex)
	w.streamKeyRegex.Wrapping = fyne.TextWrapBreak
	w.streamKeyRegex.Resize(fyne.NewSize(w.streamKeyRegex.Size().Width, 50))

	// 创建OBS配置路径输入框
	w.obsConfigPath = widget.NewEntry()
	w.obsConfigPath.SetText(config.CurrentSettings.OBSConfigPath)
	w.obsConfigPath.SetPlaceHolder("请选择OBS配置文件路径 (service.json)")

	// 创建浏览按钮
	browseBtn := widget.NewButtonWithIcon("浏览", theme.FolderOpenIcon(), func() {
		w.browseOBSConfig()
	})

	// 创建自动检测按钮
	autoDetectBtn := widget.NewButtonWithIcon("自动检测", theme.SearchIcon(), func() {
		w.autoDetectOBSConfig()
	})

	// 创建保存和取消按钮
	saveBtn := widget.NewButtonWithIcon("保存", theme.DocumentSaveIcon(), func() {
		w.saveSettings(alreadyCheck)
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("取消", func() {
		w.window.Close()
	})

	// 创建恢复默认配置按钮
	resetBtn := widget.NewButtonWithIcon("恢复默认配置", theme.HistoryIcon(), func() {
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
				w.serverRegex.SetText(config.DefaultSettings.ServerRegex)
				w.streamKeyRegex.SetText(config.DefaultSettings.StreamKeyRegex)
				w.obsConfigPath.SetText(config.DefaultSettings.OBSConfigPath)
				alreadyCheck = nil // 清空已选网卡

				// 更新当前设置为默认设置
				config.CurrentSettings = config.DefaultSettings

				dialog.ShowInformation("成功", "已恢复默认配置", w.window)
			}
		}, w.window)
	})
	resetBtn.Importance = widget.WarningImportance

	// 创建网卡列表容器
	networkCard := container.NewBorder(widget.NewCard("网卡选择", "选择要监听的网卡", nil),
		nil, nil, nil, container.NewScroll(w.networkList))
	networkCard.Resize(fyne.NewSize(600, 200))

	// 创建OBS配置路径容器
	obsPathContainer := container.NewBorder(nil, nil, nil,
		container.NewHBox(browseBtn, autoDetectBtn), w.obsConfigPath)

	// 创建配置容器
	configCard := widget.NewCard("配置设置", "",
		container.NewVBox(
			widget.NewForm(
				widget.NewFormItem("服务器地址正则", w.serverRegex),
				widget.NewFormItem("推流码正则", w.streamKeyRegex),
				widget.NewFormItem("OBS配置文件路径", obsPathContainer),
			),
		),
	)
	configCard.Resize(fyne.NewSize(600, 200))

	// 创建按钮容器，添加恢复默认配置按钮
	buttonContainer := container.NewHBox(
		saveBtn,
		resetBtn,
		cancelBtn,
	)
	buttonContainer.Resize(fyne.NewSize(600, 50))

	// 设置内容
	w.window.SetContent(container.NewBorder(configCard, buttonContainer, nil, nil, networkCard))
}

func (w *SettingsWindow) saveSettings(checks []string) {
	// 创建新的设置
	newSettings := config.Settings{
		NetworkInterfaces: checks,
		ServerRegex:       strings.TrimSpace(w.serverRegex.Text),
		StreamKeyRegex:    strings.TrimSpace(w.streamKeyRegex.Text),
		OBSConfigPath:     strings.TrimSpace(w.obsConfigPath.Text),
	}

	// 保存设置
	if err := config.SaveSettings(newSettings); err != nil {
		dialog.ShowError(err, w.window)
		return
	}

	// 更新当前设置
	config.CurrentSettings = newSettings

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
