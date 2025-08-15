package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"tiktok_tool/config"
	"tiktok_tool/lkit"
	"tiktok_tool/llog"
)

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
	liveCompanionPath := strings.TrimSpace(config.GetConfig().PathSettings.LiveCompanionPath)

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

	llog.Debug("直播伴侣启动命令已发送")
	return nil
}

// simulateClickStartLive 使用auto.exe模拟点击开始直播按钮
func (w *MainWindow) simulateClickStartLive() error {
	success, err := lkit.BringWindowToFront("直播伴侣")
	if err != nil || !success {
		return fmt.Errorf("置顶直播伴侣窗口失败: %v", err)
	}

	autoExePath := strings.TrimSpace(config.GetConfig().PathSettings.PluginScriptPath)
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

// closeLiveCompanionForAuto 为自动流程关闭直播伴侣
func (w *MainWindow) closeLiveCompanionForAuto() error {
	success, err := lkit.BringWindowToFront("直播伴侣")
	if err != nil || !success {
		return fmt.Errorf("置顶直播伴侣窗口失败: %v", err)
	}

	autoExePath := strings.TrimSpace(config.GetConfig().PathSettings.PluginScriptPath)
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

	time.Sleep(50 * time.Millisecond)

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
