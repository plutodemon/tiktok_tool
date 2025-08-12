package lkit

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
)

var panicErrorLogName = "panic_error.log"

func InitCrashLog() {
	currentDir, err := os.Getwd()
	if err == nil {
		panicErrorLogName = filepath.Join(currentDir, panicErrorLogName)
	}
	_ = os.Remove(panicErrorLogName)
}

func CrashLog() {
	r := recover()
	if r == nil {
		return
	}

	file, err := os.Create(panicErrorLogName)
	if err != nil {
		fmt.Println("创建崩溃日志文件失败: ", err)
		return
	}

	defer func() {
		_ = file.Close()
	}()

	_, _ = file.WriteString(fmt.Sprintf("程序发生严重错误: %v\n堆栈信息:\n%s", r, debug.Stack()))
}
