package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/andreykaipov/goobs"
	goobsCfg "github.com/andreykaipov/goobs/api/requests/config"
	"github.com/andreykaipov/goobs/api/typedefs"

	"tiktok_tool/config"
	"tiktok_tool/lkit"
	"tiktok_tool/llog"
)

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
	obsPath := strings.TrimSpace(config.GetConfig().PathSettings.OBSLaunchPath)
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
	obsConfigPath := strings.TrimSpace(config.GetConfig().PathSettings.OBSConfigPath)
	if obsConfigPath == "" {
		w.NewInfoDialog("提示", "请先在设置中配置OBS配置文件路径")
		return
	}

	// 检查OBS是否正在运行
	pid := isOBSRunning()
	for pid != -1 {
		// todo
		// if config.GetConfig().BaseSettings.OBSWsIp != "" {
		// 	err := w.writeOBSCfgByWebSocket()
		// 	if err != nil {
		// 		w.NewErrorDialog(fmt.Errorf("通过WebSocket导入OBS配置失败：%v", err))
		// 		return
		// 	}
		// }
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

func (w *MainWindow) writeOBSCfgByWebSocket() error {
	cfg := config.GetConfig().BaseSettings
	client, err := goobs.New(lkit.GetAddr(cfg.OBSWsIp, cfg.OBSWsPort), goobs.WithPassword(cfg.OBSWsPassword))
	if err != nil {
		return fmt.Errorf("连接OBS WebSocket失败: %v", err)
	}
	defer client.Disconnect()

	serviceType := "rtmp_custom"
	settings := &typedefs.StreamServiceSettings{
		Server: w.serverAddr.Text,
		Key:    w.streamKey.Text,
	}
	_, err = client.Config.SetStreamServiceSettings(&goobsCfg.SetStreamServiceSettingsParams{
		StreamServiceType:     &serviceType,
		StreamServiceSettings: settings,
	})
	if err != nil {
		return fmt.Errorf("设置OBS推流服务失败: %v", err)
	}

	return nil
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

	return WriteOBSConfig(strings.TrimSpace(config.GetConfig().PathSettings.OBSConfigPath), serverAddr, streamKey)
}
