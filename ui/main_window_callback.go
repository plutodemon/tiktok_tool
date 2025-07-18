package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

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
	}
	if cfg.OBSConfigPath != "" {
		w.importOBSBtn.Enable()
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
	err = cmd.Start()
	if err != nil {
		w.NewErrorDialog(err)
		return
	}

	w.app.Quit()
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

// handleImportOBS 处理导入OBS配置
func (w *MainWindow) handleImportOBS() {
	// 检查是否有推流信息
	serverAddr := strings.TrimSpace(w.serverAddr.Text)
	streamKey := strings.TrimSpace(w.streamKey.Text)

	streamKey = "aaaa"
	serverAddr = "bbbb"

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
	// 检查直播伴侣路径是否设置
	liveCompanionPath := strings.TrimSpace(config.GetConfig().LiveCompanionPath)
	if liveCompanionPath == "" {
		w.NewInfoDialog("提示", "请先在设置中配置直播伴侣启动路径")
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(liveCompanionPath); os.IsNotExist(err) {
		w.NewErrorDialog(fmt.Errorf("直播伴侣文件不存在：%s", liveCompanionPath))
		return
	}

	// 检查OBS是否正在运行
	if pid := isLiveCompanionRunning(); pid != -1 {
		w.NewInfoDialog("直播伴侣正在运行", "检测到直播伴侣已经正在运行！\n请勿重复运行直播伴侣")
		return
	}

	// 检查当前权限状态
	if lkit.IsAdmin {
		// 已经是管理员权限，直接启动
		cmd := exec.Command(liveCompanionPath)
		err := cmd.Start()
		if err != nil {
			w.NewErrorDialog(fmt.Errorf("启动直播伴侣失败：%v", err))
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
					w.NewErrorDialog(fmt.Errorf("启动直播伴侣失败：%v", err))
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
		w.NewInfoDialog("提示", "请先在设置中配置OBS启动路径")
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(obsPath); os.IsNotExist(err) {
		w.NewErrorDialog(fmt.Errorf("OBS文件不存在：%s", obsPath))
		return
	}

	// 检查OBS是否正在运行
	if pid := isOBSRunning(); pid != -1 {
		w.NewInfoDialog("OBS正在运行", "检测到OBS已经正在运行！\n请勿重复运行OBS")
		return
	}

	// 获取OBS安装目录作为工作目录
	obsDir := filepath.Dir(obsPath)

	// 启动OBS，设置正确的工作目录
	cmd := exec.Command(obsPath)
	cmd.Dir = obsDir // 设置工作目录为OBS安装目录
	err := cmd.Start()
	if err != nil {
		w.NewErrorDialog(fmt.Errorf("启动OBS失败：%v", err))
		return
	}

	w.status.SetText("OBS已启动")
}
