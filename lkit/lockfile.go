package lkit

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/nightlyone/lockfile"

	"tiktok_tool/config"
)

// 全局变量，用于在程序退出时释放锁
var appLock lockfile.Lockfile

func EnsureSingleInstance() error {
	// 使用当前工作目录作为锁文件存放位置
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前工作目录失败: %v", err)
	}

	// 锁文件路径 - 使用点开头使其成为隐藏文件
	lockPath := filepath.Join(currentDir, ".tiktok.lock")

	// 创建锁
	lock, err := lockfile.New(lockPath)
	if err != nil {
		return fmt.Errorf("创建锁文件失败: %v", err)
	}

	// 尝试获取锁
	if err := lock.TryLock(); err != nil {
		// 如果锁已存在，检查进程是否还在运行
		if errors.Is(err, lockfile.ErrBusy) {
			success, err := BringWindowToFront("抖音直播推流配置抓取")
			if err != nil || !success {
				return fmt.Errorf("置顶抖音直播推流配置抓取窗口失败: %v", err)
			}
			return config.AlreadyTop
		}
		return fmt.Errorf("获取锁失败: %v", err)
	}

	if runtime.GOOS == "windows" {
		filenameW, err := syscall.UTF16PtrFromString(lockPath)
		if err != nil {
			return err
		}
		if err := syscall.SetFileAttributes(filenameW, syscall.FILE_ATTRIBUTE_HIDDEN); err != nil {
			return fmt.Errorf("设置锁文件为隐藏失败: %v", err)
		}
	}

	// 保存锁对象以便后续释放
	appLock = lock
	return nil
}

func CleanupLock() {
	if appLock != "" {
		// 释放锁
		_ = appLock.Unlock()

		// 删除锁文件
		lockPath := string(appLock)
		_ = os.Remove(lockPath)
	}
}
