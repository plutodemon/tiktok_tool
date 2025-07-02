package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"tiktok_tool/lkit"
	"tiktok_tool/llog"
	"time"
	"unsafe"
)

func FindFileInAllDrives(fileName string) string {
	// 获取所有可用的盘符
	drives := getAllDrives()
	if len(drives) == 0 {
		return ""
	}

	// 创建上下文用于取消搜索
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultChan := make(chan string)

	// 为每个盘符启动一个goroutine进行搜索
	for _, drive := range drives {
		lkit.SafeGo(func() {
			if result := searchFileInDrive(ctx, drive, fileName); result != "" {
				select {
				case resultChan <- result:
					cancel() // 找到文件后取消其他搜索
				default:
				}
			}
		})
	}

	select {
	case result := <-resultChan:
		return result
	case <-time.After(15 * time.Second):
		cancel()
		return ""
	}
}

func getAllDrives() []string {
	var drives []string

	// 使用Windows API获取逻辑驱动器
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getLogicalDrives := kernel32.NewProc("GetLogicalDrives")

	ret, _, _ := getLogicalDrives.Call()
	if ret == 0 {
		return drives
	}

	// 检查每个可能的驱动器字母（A-Z）
	for i := 0; i < 26; i++ {
		if ret&(1<<uint(i)) != 0 {
			driveLetter := string(rune('A' + i))
			drivePath := driveLetter + ":\\"

			// 检查驱动器是否可访问
			if isDriveAccessible(drivePath) {
				drives = append(drives, drivePath)
			}
		}
	}

	return drives
}

func isDriveAccessible(drivePath string) bool {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getDriveType := kernel32.NewProc("GetDriveTypeW")

	drivePathPtr, _ := syscall.UTF16PtrFromString(drivePath)
	driveType, _, _ := getDriveType.Call(uintptr(unsafe.Pointer(drivePathPtr)))

	// 只搜索固定磁盘和可移动磁盘
	// DRIVE_FIXED = 3, DRIVE_REMOVABLE = 2
	return driveType == 3 || driveType == 2
}

func searchFileInDrive(ctx context.Context, searchPath, fileName string) string {
	var result string

	err := filepath.WalkDir(searchPath, func(path string, d os.DirEntry, err error) error {
		// 检查是否被取消
		select {
		case <-ctx.Done():
			return filepath.SkipDir
		default:
		}

		if err != nil {
			// 忽略权限错误
			if os.IsPermission(err) {
				return nil
			}
			return err
		}
		if !d.IsDir() && strings.EqualFold(d.Name(), fileName) {
			llog.Info("找到文件：", path)
			result = path
		}
		return nil
	})

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "搜索 %s 盘时出错: %v\n", searchPath, err)
	}

	return result
}
