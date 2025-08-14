package ui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"tiktok_tool/capture"
	"tiktok_tool/config"
	"tiktok_tool/lkit"
	"tiktok_tool/llog"
)

func (w *MainWindow) resetCaptureBtn() {
	w.captureBtn.SetText("开始抓包")
	w.captureBtn.Importance = widget.HighImportance
	w.captureBtn.SetIcon(theme.MediaPlayIcon())
	w.captureBtn.Refresh()

	w.restartBtn.Enable()
	w.restartBtn.Importance = widget.LowImportance
	w.restartBtn.Refresh()

	w.settingBtn.Enable()
	w.settingBtn.Importance = widget.LowImportance
	w.settingBtn.Refresh()

	cfg := config.GetConfig().PathSettings
	if cfg.OBSConfigPath != "" {
		w.importOBSBtn.Enable()
		w.importOBSBtn.Refresh()
	} else {
		w.importOBSBtn.Disable()
		w.importOBSBtn.Refresh()
	}

	w.autoBtn.SetIcon(TikTokIconResource)
	w.autoBtn.Enable()
	w.autoBtn.Refresh()
}

func (w *MainWindow) handleCapture() {
	if !capture.IsCapturing {
		// 开始抓包
		capture.IsCapturing = true
		capture.StopCapture = make(chan struct{})

		// 更改按钮样式为停止状态
		w.captureBtn.SetText("停止抓包")
		w.captureBtn.Importance = widget.DangerImportance
		w.captureBtn.SetIcon(theme.MediaPauseIcon())
		w.captureBtn.Refresh()

		w.restartBtn.Disable()
		w.settingBtn.Disable()
		w.importOBSBtn.Disable()
		w.autoBtn.Disable()
		w.autoBtn.SetIcon(TikTokIconResourceDis)
		w.autoBtn.Refresh()

		// 清空数据
		w.serverAddr.SetText("")
		w.streamKey.SetText("")
		w.ipAddr.SetText("")

		w.status.SetText("正在抓包...")

		llog.Debug("服务器地址正则表达式: ", config.GetConfig().BaseSettings.ServerRegex)
		llog.Debug("推流码正则表达式: ", config.GetConfig().BaseSettings.StreamKeyRegex)

		capture.StartCapture(
			func(server string) {
				fyne.Do(func() {
					w.serverAddr.SetText(server)
				})
			},
			func(key string) {
				fyne.Do(func() {
					w.streamKey.SetText(key)
				})
			},
			func(ip string) {
				fyne.Do(func() {
					w.ipAddr.SetText(ip)
				})
			},
			func(err error) {
				fyne.Do(func() {
					w.status.SetText("错误: " + err.Error())
				})
			},
			func() {
				fyne.Do(func() {
					w.status.SetText("已停止抓包")
					w.resetCaptureBtn()
				})
			},
		)
	} else {
		// 停止抓包
		capture.StopCapturing()
		fyne.Do(func() {
			w.status.SetText("已停止抓包")
			w.resetCaptureBtn()
		})
	}
}

// restartApp 重启应用
func (w *MainWindow) restartApp() {
	if capture.IsCapturing {
		capture.StopCapturing()
	}

	exe, err := os.Executable()
	if err != nil {
		w.NewErrorDialog(err)
		return
	}

	cmd := exec.Command(exe)

	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}

	err = cmd.Start()
	if err != nil {
		w.NewErrorDialog(err)
		return
	}

	time.AfterFunc(100*time.Millisecond, func() {
		fyne.DoAndWait(func() {
			w.app.Quit()
		})
	})
}

// handleWindowClose 处理窗口关闭事件
// 根据配置决定是最小化到托盘还是退出程序
func (w *MainWindow) handleWindowClose() {
	if config.GetConfig().BaseSettings.MinimizeOnClose {
		w.window.Hide()
		return
	}
	w.window.Close()
}

// handleAutoStart 处理一键开播功能
// 流程：启动直播伴侣 -> 开始抓包 -> 模拟点击开始直播 -> 获取推流信息 -> 导入OBS -> 启动OBS -> 关闭直播伴侣
func (w *MainWindow) handleAutoStart() {
	// 检查所有必要的配置
	if err := w.validateAutoStartConfig(); err != nil {
		w.NewErrorDialog(err)
		return
	}

	// 显示确认对话框
	confirmDialog := *w.NewConfirmDialog("一键开播确认",
		"即将执行一键开播流程：\n\n启动直播伴侣 -> 开始抓包 -> 模拟点击开始直播 -> 获取推流信息 ->"+
			" 导入OBS配置 -> 启动OBS -> 关闭直播伴侣\n\n是否继续？",
		func(confirmed bool) {
			if !confirmed {
				return
			}
			w.executeAutoStartFlow()
		})
	confirmDialog.SetDismissText("取消")
	confirmDialog.SetConfirmText("开始")
}

// validateAutoStartConfig 验证一键开播所需的配置
func (w *MainWindow) validateAutoStartConfig() error {
	cfg := config.GetConfig().PathSettings

	// 检查直播伴侣路径
	if strings.TrimSpace(cfg.LiveCompanionPath) == "" {
		return fmt.Errorf("请先在设置中配置直播伴侣启动路径")
	}
	if _, err := os.Stat(cfg.LiveCompanionPath); os.IsNotExist(err) {
		return fmt.Errorf("直播伴侣文件不存在：%s", cfg.LiveCompanionPath)
	}

	// 检查OBS路径
	if strings.TrimSpace(cfg.OBSLaunchPath) == "" {
		return fmt.Errorf("请先在设置中配置OBS启动路径")
	}
	if _, err := os.Stat(cfg.OBSLaunchPath); os.IsNotExist(err) {
		return fmt.Errorf("OBS文件不存在：%s", cfg.OBSLaunchPath)
	}

	// 检查OBS配置路径
	if strings.TrimSpace(cfg.OBSConfigPath) == "" {
		return fmt.Errorf("请先在设置中配置OBS配置文件路径")
	}
	if _, err := os.Stat(cfg.OBSConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("OBS配置文件不存在：%s", cfg.OBSConfigPath)
	}

	// 检查auto.exe脚本路径
	if strings.TrimSpace(cfg.PluginScriptPath) == "" {
		return fmt.Errorf("请先在设置中配置自动化脚本路径")
	}
	if _, err := os.Stat(cfg.PluginScriptPath); os.IsNotExist(err) {
		return fmt.Errorf("自动化脚本文件不存在：%s", cfg.PluginScriptPath)
	}

	// 检查是否已有程序在运行
	if pid := isOBSRunning(); pid != -1 {
		return fmt.Errorf("OBS已在运行，请先关闭后再使用一键开播")
	}
	// if pid := isLiveCompanionRunning(); pid != -1 {
	// 	return fmt.Errorf("直播伴侣已在运行，请先关闭后再使用一键开播")
	// }

	// 检查是否正在抓包
	if capture.IsCapturing {
		return fmt.Errorf("当前正在抓包，请先停止抓包后再使用一键开播")
	}

	return nil
}

// executeAutoStartFlow 执行一键开播流程
func (w *MainWindow) executeAutoStartFlow() {
	// 创建进度对话框
	progressLabel := widget.NewLabel("正在执行一键开播流程...")
	progressBar := widget.NewProgressBar()
	progressBar.SetValue(0.0)

	progressDialog := w.NewCustomDialog("一键开播", "", container.NewVBox(
		progressLabel,
		progressBar,
	))
	closeButton := widget.NewButton("关闭", func() {
		progressDialog.Hide()
	})
	closeButton.Disable()
	progressDialog.SetButtons([]fyne.CanvasObject{closeButton})

	onSuccess := func() {
		capture.StopCapturing()
		closeButton.Enable()
		closeButton.Refresh()
		w.autoBtn.Disable()
		w.autoBtn.SetIcon(TikTokIconResourceDis)
		w.autoBtn.Refresh()
		w.status.SetText("一键开播流程已完成！")
	}

	// 在后台执行流程
	lkit.SafeGo(func() {
		w.autoStart(progressDialog, progressLabel, progressBar, onSuccess)
	})
}

func (w *MainWindow) autoStart(progressDialog *dialog.CustomDialog, progressLabel *widget.Label, progressBar *widget.ProgressBar, onSuccess func()) {
	var progressError error
	defer func() {
		if progressError == nil {
			return
		}
		capture.StopCapturing()
		fyne.Do(func() {
			progressDialog.Hide()
			w.NewErrorDialog(progressError)
		})
	}()

	// 启动直播伴侣
	fyne.Do(func() {
		progressLabel.SetText("正在启动直播伴侣...")
		progressBar.SetValue(1.0 / 7.0)
	})
	if err := w.startLiveCompanion(false); err != nil {
		progressError = err
		return
	}

	time.Sleep(1 * time.Second)

	// 开始抓包
	fyne.Do(func() {
		progressLabel.SetText("正在抓包...")
		progressBar.SetValue(2.0 / 7.0)
	})

	onGetAll := make(chan struct{})
	channelClosed := false
	if err := w.startCaptureForAuto(func() {
		if !channelClosed {
			channelClosed = true
			close(onGetAll)
		}
	}); err != nil {
		progressError = err
		return
	}

	// 模拟点击开始直播
	fyne.Do(func() {
		progressLabel.SetText("正在模拟点击开始直播...")
		progressBar.SetValue(3.0 / 7.0)
	})

	if err := w.simulateClickStartLive(); err != nil {
		progressError = err
		return
	}

	timeout := time.After(20 * time.Second)
	select {
	case <-onGetAll:
		llog.Debug("成功获取到推流信息: ", w.serverAddr.Text, w.streamKey.Text)
	case <-timeout:
		progressError = fmt.Errorf("获取推流信息超时，请检查配置文件或检查网络连接或重试")
		return
	}

	time.Sleep(500 * time.Millisecond)

	// 导入OBS配置
	fyne.Do(func() {
		progressLabel.SetText("正在导入OBS配置...")
		progressBar.SetValue(4.0 / 7.0)
	})
	if err := w.importOBSConfigForAuto(); err != nil {
		progressError = fmt.Errorf("导入OBS配置失败：%v", err)
		return
	}

	// 启动OBS
	fyne.Do(func() {
		progressLabel.SetText("正在启动OBS...")
		progressBar.SetValue(5.0 / 7.0)
	})
	if err := w.startOBS(false); err != nil {
		progressError = err
		return
	}

	// 关闭直播伴侣
	fyne.Do(func() {
		progressLabel.SetText("正在关闭直播伴侣...")
		progressBar.SetValue(6.0 / 7.0)
	})
	if err := w.closeLiveCompanionForAuto(); err != nil {
		progressError = err
		return
	}

	// 完成
	fyne.Do(func() {
		progressLabel.SetText("一键开播完成！")
		progressBar.SetValue(1)
		onSuccess()
	})
}

// startCaptureForAuto 为自动流程开始抓包
func (w *MainWindow) startCaptureForAuto(onGetAll func()) error {
	// 开始抓包
	capture.IsCapturing = true
	capture.StopCapture = make(chan struct{})

	// 清空数据
	fyne.Do(func() {
		w.serverAddr.SetText("")
		w.streamKey.SetText("")
	})

	var err error
	// 启动抓包
	capture.StartCapture(
		func(server string) {
			fyne.DoAndWait(func() {
				w.serverAddr.SetText(server)
			})
		},
		func(streamKey string) {
			fyne.DoAndWait(func() {
				w.streamKey.SetText(streamKey)
			})
		},
		func(ip string) {
			fyne.DoAndWait(func() {
				w.ipAddr.SetText(ip)
			})
		},
		func(err error) {
			err = fmt.Errorf("抓包过程中发生错误: %v", err)
		},
		onGetAll,
	)

	return err
}
