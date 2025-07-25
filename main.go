package main

import (
	_ "embed"
	"fmt"

	"tiktok_tool/config"
	"tiktok_tool/lkit"
	"tiktok_tool/llog"
	"tiktok_tool/ui"
)

func main() {
	lkit.InitCrashLog()
	defer lkit.CrashLog()

	if err := config.LoadConfig(); err != nil {
		panic(fmt.Sprintf("初始化配置失败: %v", err))
	}

	if err := llog.Init(config.GetConfig().LogConfig); err != nil {
		panic(fmt.Sprintf("初始化日志系统失败: %v", err))
	}
	defer llog.Cleanup()

	// 设置全局panic处理
	defer llog.HandlePanic()

	ui.NewMainWindow()
}
