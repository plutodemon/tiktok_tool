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

	"tiktok_tool/config"
	"tiktok_tool/lkit"
)

// createScriptTab 创建脚本设置标签页
func (w *SettingsWindow) createScriptTab() fyne.CanvasObject {
	// 创建插件脚本浏览按钮
	browsePluginScriptBtn := widget.NewButtonWithIcon("浏览", theme.FolderOpenIcon(), w.browsePluginScriptConfig)

	// 创建检测插件按钮
	detectBtn := widget.NewButtonWithIcon("检测插件", theme.SearchIcon(), w.detectPlugin)

	// 创建清空插件脚本路径按钮
	clearScriptBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		w.pluginScriptDownloadBtn.Enable()
		w.pluginScriptDownloadBtn.Importance = widget.SuccessImportance
		w.pluginScriptDownloadBtn.Refresh()
		w.pluginScriptPath.SetText("")
	})

	// 创建下载按钮
	if config.GetConfig().PluginScriptPath != "" {
		w.pluginScriptDownloadBtn.Disable()
	} else {
		w.pluginScriptDownloadBtn.Importance = widget.SuccessImportance
	}

	// 创建清空插件按钮
	clearBtn := widget.NewButtonWithIcon("清空插件", theme.DeleteIcon(), w.clearPlugin)
	clearBtn.Importance = widget.DangerImportance

	// 创建插件脚本路径容器
	pluginScriptPathContainer := container.NewBorder(nil, nil, nil,
		container.NewHBox(browsePluginScriptBtn, detectBtn, clearScriptBtn), w.pluginScriptPath)

	// 创建插件管理按钮容器
	pluginManageContainer := container.New(
		layout.NewGridLayout(2),
		w.pluginScriptDownloadBtn,
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
				w.pluginScriptDownloadBtn.Disable()
				w.pluginScriptDownloadBtn.Refresh()
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
			w.pluginScriptDownloadBtn.Enable()
			w.pluginScriptDownloadBtn.Importance = widget.SuccessImportance
			w.pluginScriptDownloadBtn.Refresh()
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
	progressDialog := w.NewCustomWithoutButtons("下载auto.exe", container.NewVBox(
		progressLabel,
		progressBar,
	))

	// 在后台执行下载
	lkit.SafeGo(func() {
		// 确保plugin目录存在
		pluginDir := "plugin"
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			fyne.Do(func() {
				progressDialog.Hide()
				w.NewErrorDialog(fmt.Errorf("创建plugin目录失败: %v", err))
			})
			return
		}

		// 下载文件
		downloadURL := "https://github.com/plutodemon/py_win_auto/releases/download/v0.1.1/auto.exe"
		filePath := filepath.Join(pluginDir, "auto.exe")

		fyne.Do(func() {
			progressLabel.SetText("正在下载...")
		})
		resp, err := http.Get(downloadURL)
		if err != nil {
			fyne.Do(func() {
				progressDialog.Hide()
				w.NewErrorDialog(fmt.Errorf("下载失败: %v", err))
			})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fyne.Do(func() {
				progressDialog.Hide()
				w.NewErrorDialog(fmt.Errorf("下载失败: HTTP %d", resp.StatusCode))
			})
			return
		}

		// 创建目标文件
		file, err := os.Create(filePath)
		if err != nil {
			fyne.Do(func() {
				progressDialog.Hide()
				w.NewErrorDialog(fmt.Errorf("创建文件失败: %v", err))
			})
			return
		}
		defer file.Close()

		// 复制文件内容
		fyne.Do(func() {
			progressLabel.SetText("正在保存文件...")
			progressBar.SetValue(0.7)
		})
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			fyne.Do(func() {
				progressDialog.Hide()
				w.NewErrorDialog(fmt.Errorf("保存文件失败: %v", err))
			})
			return
		}

		// 下载完成
		fyne.Do(func() {
			progressBar.SetValue(1.0)
			progressLabel.SetText("下载完成")
		})
		time.Sleep(500 * time.Millisecond)
		fyne.Do(func() {
			progressDialog.Hide()

			// 自动设置路径
			w.pluginScriptDownloadBtn.Disable()
			w.pluginScriptDownloadBtn.Refresh()
			w.pluginScriptPath.SetText(filePath)

			// 显示成功消息
			w.NewInfoDialog("下载成功", fmt.Sprintf("auto.exe已下载到:\n%s\n\n路径已自动设置到脚本路径中。", filePath))
		})
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
		w.pluginScriptDownloadBtn.Disable()
		w.pluginScriptDownloadBtn.Refresh()
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
