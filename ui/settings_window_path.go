package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"tiktok_tool/config"
	"tiktok_tool/lkit"
)

// createPathTab 创建其他设置标签页
func (w *SettingsWindow) createPathTab(alreadyCheck *[]string) fyne.CanvasObject {
	// 创建OBS启动路径浏览按钮
	browseOBSLaunchBtn := widget.NewButtonWithIcon("浏览", theme.FolderOpenIcon(), w.browseOBSLaunchConfig)

	// 创建OBS启动路径自动检测按钮
	autoDetectOBSLaunchBtn := widget.NewButtonWithIcon("自动检测", theme.SearchIcon(), w.autoDetectOBSLaunchConfig)

	// 创建清空OBS启动路径按钮
	clearOBSLaunchBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		w.obsLaunchPath.SetText("")
	})

	// 创建OBS配置文件浏览按钮
	browseBtn := widget.NewButtonWithIcon("浏览", theme.FolderOpenIcon(), w.browseOBSConfig)

	// 创建OBS配置文件自动检测按钮
	autoDetectBtn := widget.NewButtonWithIcon("自动检测", theme.SearchIcon(), w.autoDetectOBSConfig)

	// 创建清空OBS配置文件路径按钮
	clearOBSConfigBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		w.obsConfigPath.SetText("")
	})

	// 创建直播伴侣浏览按钮
	browseLiveCompanionBtn := widget.NewButtonWithIcon("浏览", theme.FolderOpenIcon(), w.browseLiveCompanionConfig)

	// 创建直播伴侣自动检测按钮
	autoDetectLiveCompanionBtn := widget.NewButtonWithIcon("自动检测", theme.SearchIcon(), w.autoDetectLiveCompanionConfig)

	// 创建清空直播伴侣路径按钮
	clearLiveCompanionBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		w.liveCompanionPath.SetText("")
	})

	// 创建OBS启动路径容器
	obsLaunchPathContainer := container.NewBorder(nil, nil, nil,
		container.NewHBox(browseOBSLaunchBtn, autoDetectOBSLaunchBtn, clearOBSLaunchBtn), w.obsLaunchPath)

	// 创建OBS配置路径容器
	obsPathContainer := container.NewBorder(nil, nil, nil,
		container.NewHBox(browseBtn, autoDetectBtn, clearOBSConfigBtn), w.obsConfigPath)

	// 创建直播伴侣路径容器
	liveCompanionPathContainer := container.NewBorder(nil, nil, nil,
		container.NewHBox(browseLiveCompanionBtn, autoDetectLiveCompanionBtn, clearLiveCompanionBtn), w.liveCompanionPath)

	// 创建表单
	otherForm := widget.NewForm(
		widget.NewFormItem("OBS启动路径", obsLaunchPathContainer),
		widget.NewFormItem("OBS配置文件路径", obsPathContainer),
		widget.NewFormItem("直播伴侣启动路径", liveCompanionPathContainer),
	)

	// 创建恢复默认配置按钮
	resetBtn := widget.NewButtonWithIcon("恢复默认配置", theme.HistoryIcon(), func() {
		w.resetToDefaults()
	})

	// 添加说明文本
	otherHelp := widget.NewRichTextFromMarkdown("### 配置说明\n\n" +
		"* **OBS启动路径**：OBS Studio的可执行文件路径，用于快速启动OBS\n" +
		"* **OBS配置文件路径**：OBS Studio的配置文件路径，用于导入OBS推流设置\n" +
		"* **直播伴侣启动路径**：抖音直播伴侣的可执行文件路径，用于快速启动直播伴侣\n")

	// 创建容器
	return container.NewVBox(
		otherForm,
		layout.NewSpacer(),
		otherHelp,
		widget.NewSeparator(),
		resetBtn,
		widget.NewSeparator(),
	)
}

// resetToDefaults 重置为默认设置
func (w *SettingsWindow) resetToDefaults() {
	w.NewConfirmDialog("确认", "确定要恢复默认配置吗？", func(ok bool) {
		if !ok {
			return
		}

		configPath := filepath.Join(config.CfgFilePath, config.CfgFileName)
		if _, err := os.Stat(configPath); err == nil {
			// 配置文件在 config 目录中
			if err := os.Remove(configPath); err != nil {
				w.NewErrorDialog(fmt.Errorf("删除配置文件失败: %v", err))
				return
			}
		} else {
			// 检查当前目录
			configPath = filepath.Join(config.CfgFileName)
			if _, err = os.Stat(configPath); err == nil {
				if err = os.Remove(configPath); err != nil {
					w.NewErrorDialog(fmt.Errorf("删除配置文件失败: %v", err))
					return
				}
			}
		}

		w.close()
		w.saveCallback("已恢复默认配置，请重启软件以应用更改")
	})
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
		if path := config.GetConfig().PathSettings.OBSConfigPath; path != "" {
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
	detectedPath := GetDefaultOBSConfigPath()
	if detectedPath == "" {
		w.NewInfoDialog("检测结果", "未找到OBS配置文件，请确保OBS Studio已安装并至少运行过一次")
		return
	}

	w.obsConfigPath.SetText(detectedPath)
	w.NewInfoDialog("检测成功", fmt.Sprintf("已自动检测到OBS配置文件：\n%s", detectedPath))
}

// GetDefaultOBSConfigPath 获取默认的OBS配置文件路径(service.json)
func GetDefaultOBSConfigPath() string {
	// Windows系统下OBS service.json配置文件的常见路径
	possiblePaths := []string{
		filepath.Join(os.Getenv("APPDATA"), "obs-studio", "basic", "profiles", "Untitled", "service.json"),
		filepath.Join(os.Getenv("APPDATA"), "obs-studio", "basic", "profiles", "Default", "service.json"),
	}

	// 检查每个可能的路径
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 尝试查找其他配置文件
	profilesDir := filepath.Join(os.Getenv("APPDATA"), "obs-studio", "basic", "profiles")
	if entries, err := os.ReadDir(profilesDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				path := filepath.Join(profilesDir, entry.Name(), "service.json")
				if _, err := os.Stat(path); err == nil {
					return path
				}
			}
		}
	}

	return ""
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

	if path := config.GetConfig().PathSettings.LiveCompanionPath; path != "" {
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

// autoDetectLiveCompanionConfig 自动检测直播伴侣启动路径
func (w *SettingsWindow) autoDetectLiveCompanionConfig() {
	progressDialog := w.NewCustomWithoutButtons(
		"搜索中",
		container.NewCenter(widget.NewLabel("正在搜索直播伴侣安装路径，请稍候...")),
	)

	lkit.SafeGo(func() {
		result, title, info := "", "", ""
		defer func() {
			fyne.Do(func() {
				progressDialog.Hide()
				w.liveCompanionPath.SetText(result)
				w.NewInfoDialog(title, info)
			})
		}()
		result = lkit.FindFileInAllDrives("直播伴侣 Launcher.exe")
		if result == "" {
			title = "检测失败"
			info = "未找到直播伴侣安装路径。\n\n可能的原因：\n1. 抖音直播伴侣未安装\n2. 安装在非标准位置\n3. 文件名与预期不符\n\n请尝试手动浏览选择。"
			return
		}
		title = "检测成功"
		info = fmt.Sprintf("已自动检测到直播伴侣启动路径：\n%s", result)
	})
}

// browseOBSLaunchConfig 浏览选择OBS启动文件
func (w *SettingsWindow) browseOBSLaunchConfig() {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if reader == nil {
			return
		}
		defer reader.Close()

		filePath := reader.URI().Path()
		w.obsLaunchPath.SetText(filePath)
	}, w.window)

	// 设置文件过滤器
	filter := storage.NewExtensionFileFilter([]string{".exe"})
	fileDialog.SetFilter(filter)
	fileDialog.SetConfirmText("确定选择")
	fileDialog.SetDismissText("取消选择")

	if path := config.GetConfig().PathSettings.OBSLaunchPath; path != "" {
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

// autoDetectOBSLaunchConfig 自动检测OBS启动路径
func (w *SettingsWindow) autoDetectOBSLaunchConfig() {
	progressDialog := w.NewCustomWithoutButtons(
		"搜索中",
		container.NewCenter(widget.NewLabel("正在搜索OBS启动路径，请稍候...")),
	)

	resultChan := make(chan string)

	lkit.SafeGo(func() {
		resultChan <- lkit.FindFileInAllDrives("obs64.exe")
	})

	select {
	case result := <-resultChan:
		progressDialog.Hide()
		if result == "" {
			w.NewInfoDialog("检测结果", "未找到OBS启动路径。请尝试手动浏览选择。")
			return
		}
		w.obsLaunchPath.SetText(result)
		w.NewInfoDialog("检测成功", fmt.Sprintf("已自动检测到OBS启动路径：\n%s", result))
	}
}
