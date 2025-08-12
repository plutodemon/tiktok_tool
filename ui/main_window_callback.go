package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

	cfg := config.GetConfig()
	if cfg.OBSLaunchPath != "" {
		w.obsBtn.SetIcon(OBSIconResource)
		w.obsBtn.Enable()
		w.obsBtn.Refresh()
	} else {
		w.obsBtn.SetIcon(OBSIconResourceDis)
		w.obsBtn.Disable()
		w.obsBtn.Refresh()
	}

	if cfg.OBSConfigPath != "" {
		w.importOBSBtn.Enable()
		w.importOBSBtn.Refresh()
	} else {
		w.importOBSBtn.Disable()
		w.importOBSBtn.Refresh()
	}
}

func (w *MainWindow) handleCapture() {
	if !config.IsCapturing {
		// 开始抓包
		config.IsCapturing = true
		config.StopCapture = make(chan struct{})

		// 更改按钮样式为停止状态
		w.captureBtn.SetText("停止抓包")
		w.captureBtn.Importance = widget.DangerImportance
		w.captureBtn.SetIcon(theme.MediaPauseIcon())
		w.captureBtn.Refresh()

		w.restartBtn.Disable()
		w.settingBtn.Disable()
		w.importOBSBtn.Disable()
		w.obsBtn.Disable()
		w.obsBtn.SetIcon(OBSIconResourceDis)
		w.obsBtn.Refresh()

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

			keyRegex := config.GetConfig().StreamKeyRegex
			if len(keyRegex) == 0 {
				keyRegex = config.DefaultConfig.StreamKeyRegex
			}
			llog.Debug("推流码正则表达式: ", keyRegex)
		}

		capture.StartCapture(
			func(server string) {
				fyne.Do(func() {
					w.serverAddr.SetText(server)
					w.status.SetText("已找到推流服务器地址")
				})
			},
			func(key string) {
				fyne.Do(func() {
					w.streamKey.SetText(key)
					w.status.SetText("已找到推流码")
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
	if config.IsCapturing {
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

// isOBSRunning 检查OBS是否正在运行
func isOBSRunning() int32 {
	pids, err := lkit.IsProcessRunning("obs64.exe", "obs32.exe")
	if err != nil {
		llog.Error("检查OBS进程失败:", err)
		return -1
	}
	if pids[0] > 0 {
		return pids[0]
	}
	if pids[1] > 0 {
		return pids[1]
	}
	return -1
}

func isLiveCompanionRunning() int32 {
	pids, err := lkit.IsProcessRunning("直播伴侣.exe")
	if err != nil {
		llog.Error("检查直播伴侣进程失败:", err)
		return -1
	}
	if pids[0] > 0 {
		return pids[0]
	}
	return -1
}

// handleWindowClose 处理窗口关闭事件
// 根据配置决定是最小化到托盘还是退出程序
func (w *MainWindow) handleWindowClose() {
	cfg := config.GetConfig()
	if cfg.MinimizeOnClose {
		// 最小化到系统托盘
		w.window.Hide()
	} else {
		// 正常退出程序
		w.window.Close()
	}
}

// handleImportOBS 处理导入OBS配置
func (w *MainWindow) handleImportOBS() {
	// 检查是否有推流信息
	serverAddr := strings.TrimSpace(w.serverAddr.Text)
	streamKey := strings.TrimSpace(w.streamKey.Text)

	if serverAddr == "" || streamKey == "" {
		w.NewInfoDialog("提示", "请先抓取到推流服务器地址和推流码")
		return
	}

	// 检查OBS配置路径是否设置
	obsConfigPath := strings.TrimSpace(config.GetConfig().OBSConfigPath)
	if obsConfigPath == "" {
		w.NewInfoDialog("提示", "请先在设置中配置OBS配置文件路径")
		return
	}

	// 检查OBS是否正在运行
	if pid := isOBSRunning(); pid != -1 {
		closeConfirm := *w.NewConfirmDialog("OBS正在运行",
			"检测到OBS正在运行，导入配置需要先关闭OBS。\n是否要自动关闭OBS？(再次启动会有提示)\n(建议：手动关闭OBS)",
			func(confirmed bool) {
				if !confirmed {
					return
				}
				err := lkit.KillProcess(pid)
				if err == nil {
					return
				}
				w.NewErrorDialog(fmt.Errorf("关闭OBS失败：%v", err))
			})
		closeConfirm.SetDismissText("手动关闭")
		closeConfirm.SetConfirmText("自动关闭")
		return
	}

	// 确认对话框
	writeConfirm := *w.NewConfirmDialog("确认导入", "将要导入配置到以下OBS配置文件中：\n"+obsConfigPath,
		func(confirmed bool) {
			if !confirmed {
				return
			}

			// 写入OBS配置
			err := WriteOBSConfig(obsConfigPath, serverAddr, streamKey)
			if err != nil {
				w.NewErrorDialog(fmt.Errorf("导入OBS配置失败：%v", err))
				return
			}

			w.status.SetText("OBS配置导入成功")
			w.NewInfoDialog("成功", "推流配置已成功导入到OBS！")
		})
	writeConfirm.SetDismissText("取消")
	writeConfirm.SetConfirmText("导入")
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
	quit := false
	if lkit.IsAdmin == false {
		confirmDialog := *w.NewConfirmDialog("管理员权限确认",
			"启动直播伴侣需要管理员权限，系统将弹出UAC提示\n是否继续启动？",
			func(confirmed bool) {
				quit = !confirmed
			})
		confirmDialog.SetDismissText("取消")
		confirmDialog.SetConfirmText("继续")
	}

	if quit {
		return
	}

	if err := w.startLiveCompanion(true); err != nil {
		w.NewErrorDialog(err)
		return
	}

	w.status.SetText("直播伴侣启动请求已发送")
}

// startLiveCompanion 启动直播伴侣
func (w *MainWindow) startLiveCompanion(check bool) error {
	liveCompanionPath := strings.TrimSpace(config.GetConfig().LiveCompanionPath)

	// 检查路径是否为空
	if liveCompanionPath == "" {
		return fmt.Errorf("请先在设置中配置直播伴侣启动路径")
	}

	// 检查文件是否存在
	if _, err := os.Stat(liveCompanionPath); os.IsNotExist(err) {
		return fmt.Errorf("直播伴侣文件不存在：%s", liveCompanionPath)
	}

	// 检查是否已经运行
	// if pid := isLiveCompanionRunning(); check && pid != -1 {
	// 	success, err := lkit.BringWindowToFront("直播伴侣")
	// 	if err != nil || !success {
	// 		return fmt.Errorf("检测到直播伴侣已经正在运行！\n置顶直播伴侣窗口失败: %v", err)
	// 	}
	// 	return fmt.Errorf("检测到直播伴侣已经正在运行！\n请勿重复运行直播伴侣(已置顶窗口)")
	// }

	if lkit.IsAdmin {
		// 已经是管理员权限，直接启动
		cmd := exec.Command(liveCompanionPath)
		err := cmd.Start()
		if err != nil {
			return fmt.Errorf("启动直播伴侣失败：%v", err)
		}
	} else {
		// 使用PowerShell以管理员权限启动直播伴侣
		powershellCmd := fmt.Sprintf("Start-Process -FilePath '%s' -Verb RunAs", liveCompanionPath)
		cmd := exec.Command("powershell", "-Command", powershellCmd)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow: true,
		}
		err := cmd.Start()
		if err != nil {
			return fmt.Errorf("启动直播伴侣失败：%v", err)
		}
	}

	return nil
}

// handleStartOBS 处理启动OBS
func (w *MainWindow) handleStartOBS() {
	if err := w.startOBS(true); err != nil {
		w.NewErrorDialog(err)
		return
	}

	w.status.SetText("OBS启动请求已发送")
}

// startOBSForAuto 为自动流程启动OBS
func (w *MainWindow) startOBS(check bool) error {
	obsPath := strings.TrimSpace(config.GetConfig().OBSLaunchPath)
	if obsPath == "" {
		return fmt.Errorf("请先在设置中配置OBS启动路径")
	}

	// 检查文件是否存在
	if _, err := os.Stat(obsPath); os.IsNotExist(err) {
		return fmt.Errorf("OBS文件不存在：%s", obsPath)
	}

	// 检查OBS是否正在运行
	if pid := isOBSRunning(); pid != -1 {
		success, err := lkit.BringWindowToFront("OBS")
		if err != nil || !success {
			return fmt.Errorf("检测到OBS已经正在运行！\n置顶OBS窗口失败: %v", err)
		}
		if check {
			return fmt.Errorf("检测到OBS已经正在运行！\n请勿重复运行OBS(已置顶窗口)")
		}
		return nil
	}

	// 获取OBS安装目录作为工作目录
	obsDir := filepath.Dir(obsPath)

	// 启动OBS，设置正确的工作目录
	cmd := exec.Command(obsPath)
	cmd.Dir = obsDir
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("启动OBS失败：%v", err)
	}

	return nil
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
	cfg := config.GetConfig()

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
	config.IsCapturing = true
	config.StopCapture = make(chan struct{})

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
		func(err error) {
			err = fmt.Errorf("抓包过程中发生错误: %v", err)
		},
		onGetAll,
	)

	return err
}

// simulateClickStartLive 使用auto.exe模拟点击开始直播按钮
func (w *MainWindow) simulateClickStartLive() error {
	success, err := lkit.BringWindowToFront("直播伴侣")
	if err != nil || !success {
		return fmt.Errorf("置顶直播伴侣窗口失败: %v", err)
	}

	autoExePath := strings.TrimSpace(config.GetConfig().PluginScriptPath)
	args := []string{"--app", "直播伴侣", "--control", "开始直播", "--type", "Text"}

	result, err := lkit.RunAutoTool(autoExePath, args)
	if err != nil {
		return fmt.Errorf("获取开始直播按钮位置失败：%v", err)
	}
	if !result.Success {
		return fmt.Errorf("获取开始直播按钮位置失败：%s", result.Error)
	}
	err = lkit.SimulateLeftClick(result.Center.X, result.Center.Y)
	if err != nil {
		return fmt.Errorf("模拟点击开始直播按钮失败：%v", err)
	}

	return nil
}

// importOBSConfigForAuto 为自动流程导入OBS配置
func (w *MainWindow) importOBSConfigForAuto() error {
	serverAddr := strings.TrimSpace(w.serverAddr.Text)
	streamKey := strings.TrimSpace(w.streamKey.Text)

	if serverAddr == "" || streamKey == "" {
		return fmt.Errorf("推流信息不完整")
	}

	if pid := isOBSRunning(); pid != -1 {
		return fmt.Errorf("OBS正在运行，请先关闭OBS后再导入配置")
	}

	return WriteOBSConfig(strings.TrimSpace(config.GetConfig().OBSConfigPath), serverAddr, streamKey)
}

// closeLiveCompanionForAuto 为自动流程关闭直播伴侣
func (w *MainWindow) closeLiveCompanionForAuto() error {
	success, err := lkit.BringWindowToFront("直播伴侣")
	if err != nil || !success {
		return fmt.Errorf("置顶直播伴侣窗口失败: %v", err)
	}

	autoExePath := strings.TrimSpace(config.GetConfig().PluginScriptPath)
	args := []string{"--app", "直播伴侣", "--control", "关闭", "--type", "Button"}

	result, err := lkit.RunAutoTool(autoExePath, args)
	if err != nil {
		return fmt.Errorf("获取关闭按钮位置失败：%v", err)
	}

	if !result.Success {
		return fmt.Errorf("获取关闭按钮位置失败：%s", result.Error)
	}
	err = lkit.SimulateLeftClick(result.Center.X, result.Center.Y)
	if err != nil {
		return fmt.Errorf("模拟点击关闭按钮失败：%v", err)
	}

	time.Sleep(1 * time.Second)

	args = []string{"--app", "直播伴侣", "--control", "确定", "--type", "Button"}

	result, err = lkit.RunAutoTool(autoExePath, args)
	if err != nil {
		return fmt.Errorf("获取关闭按钮位置失败：%v", err)
	}

	if !result.Success {
		return fmt.Errorf("获取关闭按钮位置失败：%s", result.Error)
	}
	err = lkit.SimulateLeftClick(result.Center.X, result.Center.Y)
	if err != nil {
		return fmt.Errorf("模拟点击关闭按钮失败：%v", err)
	}

	return nil
}
