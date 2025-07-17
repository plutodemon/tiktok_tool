package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"tiktok_tool/capture"
	"tiktok_tool/config"
	"tiktok_tool/lkit"
	"tiktok_tool/llog"
)

func (w *MainWindow) handleCapture() {
	if !config.IsCapturing {
		// 开始抓包
		config.IsCapturing = true
		config.StopCapture = make(chan struct{})

		// 更改按钮样式为停止状态
		w.captureBtn.SetText("停止抓包")
		w.captureBtn.Importance = widget.DangerImportance
		w.captureBtn.SetIcon(theme.MediaPauseIcon())

		w.restartBtn.Disable()
		w.settingBtn.Disable()

		// 清空数据
		w.serverAddr.SetText("")
		w.streamKey.SetText("")

		w.status.SetText("正在抓包...")

		if config.IsDebug {
			serverRegexCompile := config.GetConfig().ServerRegex
			if len(serverRegexCompile) == 0 {
				serverRegexCompile = config.DefaultConfig.ServerRegex
			}
			llog.Debug("服务器地址正则表达式: ", serverRegexCompile)

			keyRegex := config.DefaultConfig.StreamKeyRegex
			if len(config.GetConfig().StreamKeyRegex) > 0 && config.GetConfig().StreamKeyRegex != keyRegex {
				keyRegex = config.GetConfig().StreamKeyRegex
			}
			llog.Debug("推流码正则表达式: ", keyRegex)
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
				w.resetCaptureBtn()
			},
		)
	} else {
		// 停止抓包
		capture.StopCapturing()
		w.status.SetText("已停止抓包")
		w.resetCaptureBtn()
	}
}

func (w *MainWindow) restartApp() {
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
	obsConfigPath := strings.TrimSpace(config.GetConfig().OBSConfigPath)
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
			err := WriteOBSConfig(obsConfigPath, serverAddr, streamKey)
			if err != nil {
				dialog.ShowError(fmt.Errorf("导入OBS配置失败：%v", err), w.window)
				return
			}

			w.status.SetText("OBS配置导入成功")
			dialog.ShowInformation("成功", "推流配置已成功导入到OBS！", w.window)
		}, w.window)
}

// WriteOBSConfig 将推流配置写入OBS配置文件(service.json)
func WriteOBSConfig(configPath, server, key string) error {
	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 读取JSON配置文件
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析JSON
	var cfgMap map[string]interface{}
	err = json.Unmarshal(content, &cfgMap)
	if err != nil {
		return fmt.Errorf("解析JSON配置文件失败: %v", err)
	}

	// 确保settings字段存在
	if cfgMap["settings"] == nil {
		cfgMap["settings"] = make(map[string]interface{})
	}

	// 获取settings对象
	settings, ok := cfgMap["settings"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("配置文件格式错误: settings字段不是对象")
	}

	// 更新server和key字段
	settings["server"] = server
	settings["key"] = key

	// 将修改后的配置转换回JSON
	newContent, err := json.MarshalIndent(cfgMap, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化JSON配置失败: %v", err)
	}

	// 写回文件
	err = os.WriteFile(configPath, newContent, 0644)
	if err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// handleStartLiveCompanion 处理启动直播伴侣
func (w *MainWindow) handleStartLiveCompanion() {
	// 检查直播伴侣路径是否设置
	liveCompanionPath := strings.TrimSpace(config.GetConfig().LiveCompanionPath)
	if liveCompanionPath == "" {
		dialog.ShowInformation("提示", "请先在设置中配置直播伴侣启动路径", w.window)
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(liveCompanionPath); os.IsNotExist(err) {
		dialog.ShowError(fmt.Errorf("直播伴侣文件不存在：%s", liveCompanionPath), w.window)
		return
	}

	// 检查当前权限状态
	if lkit.IsAdmin {
		// 已经是管理员权限，直接启动
		cmd := exec.Command(liveCompanionPath)
		err := cmd.Start()
		if err != nil {
			dialog.ShowError(fmt.Errorf("启动直播伴侣失败：%v", err), w.window)
			return
		}
		w.status.SetText("直播伴侣已启动（管理员权限）")
	} else {
		// 普通用户权限，显示确认对话框并以管理员权限启动
		confirmDialog := *w.NewConfirmDialog("管理员权限确认",
			"启动直播伴侣需要管理员权限，系统将弹出UAC提示\n是否继续启动？",
			func(confirmed bool) {
				if !confirmed {
					return
				}

				// 使用PowerShell以管理员权限启动直播伴侣
				powershellCmd := fmt.Sprintf("Start-Process -FilePath '%s' -Verb RunAs", liveCompanionPath)
				cmd := exec.Command("powershell", "-Command", powershellCmd)

				// 隐藏PowerShell窗口
				cmd.SysProcAttr = &syscall.SysProcAttr{
					HideWindow: true,
				}

				err := cmd.Start()
				if err != nil {
					dialog.ShowError(fmt.Errorf("启动直播伴侣失败：%v", err), w.window)
					return
				}

				w.status.SetText("直播伴侣启动请求已发送（提升权限）")
			})
		confirmDialog.SetDismissText("取消")
		confirmDialog.SetConfirmText("继续")
	}
}

// handleStartOBS 处理启动OBS
func (w *MainWindow) handleStartOBS() {
	// 检查OBS路径是否设置
	obsPath := strings.TrimSpace(config.GetConfig().OBSLaunchPath)
	if obsPath == "" {
		dialog.ShowInformation("提示", "请先在设置中配置OBS启动路径", w.window)
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(obsPath); os.IsNotExist(err) {
		dialog.ShowError(fmt.Errorf("OBS文件不存在：%s", obsPath), w.window)
		return
	}

	// 获取OBS安装目录作为工作目录
	obsDir := filepath.Dir(obsPath)

	// 启动OBS，设置正确的工作目录
	cmd := exec.Command(obsPath)
	cmd.Dir = obsDir // 设置工作目录为OBS安装目录
	err := cmd.Start()
	if err != nil {
		dialog.ShowError(fmt.Errorf("启动OBS失败：%v", err), w.window)
		return
	}

	w.status.SetText("OBS已启动")
}
