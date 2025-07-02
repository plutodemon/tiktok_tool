package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"tiktok_tool/capture"
	"tiktok_tool/config"
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
			err := config.WriteOBSConfig(obsConfigPath, serverAddr, streamKey)
			if err != nil {
				dialog.ShowError(fmt.Errorf("导入OBS配置失败：%v", err), w.window)
				return
			}

			w.status.SetText("OBS配置导入成功")
			dialog.ShowInformation("成功", "推流配置已成功导入到OBS！", w.window)
		}, w.window)
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

	// 显示确认对话框，提醒用户需要管理员权限
	dialog.ShowConfirm("管理员权限确认",
		"启动直播伴侣需要管理员权限，系统将弹出UAC提示，请点击'是'以继续。\n\n是否继续启动？",
		func(confirmed bool) {
			if !confirmed {
				return
			}

			// 使用PowerShell以管理员权限启动直播伴侣
			// 构建PowerShell命令：Start-Process -FilePath "路径" -Verb RunAs
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

			w.status.SetText("直播伴侣启动请求已发送")
		}, w.window)
}
