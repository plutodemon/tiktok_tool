package lkit

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"

	"tiktok_tool/llog"
)

var IsAdmin bool

var (
	user32                  = windows.NewLazySystemDLL("user32.dll")
	procEnumWindows         = user32.NewProc("EnumWindows")
	procGetWindowTextW      = user32.NewProc("GetWindowTextW")
	procIsWindowVisible     = user32.NewProc("IsWindowVisible")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procShowWindow          = user32.NewProc("ShowWindow")
	procSetCursorPos        = user32.NewProc("SetCursorPos")
	procSendInput           = user32.NewProc("SendInput")
)

const (
	SW_SHOW    = 5
	SW_RESTORE = 9
)

func init() {
	IsAdmin = IsRunAsAdmin()
}

// IsRunAsAdmin 检测当前程序是否以管理员身份运行
// 返回值：true表示以管理员身份运行，false表示普通用户权限
func IsRunAsAdmin() bool {
	// 加载advapi32.dll
	advapi32 := syscall.NewLazyDLL("advapi32.dll")
	getCurrentProcess := syscall.NewLazyDLL("kernel32.dll").NewProc("GetCurrentProcess")
	openProcessToken := advapi32.NewProc("OpenProcessToken")
	getTokenInformation := advapi32.NewProc("GetTokenInformation")

	// 获取当前进程句柄
	currentProcess, _, _ := getCurrentProcess.Call()

	// 打开进程令牌
	var token syscall.Handle
	ret, _, _ := openProcessToken.Call(
		currentProcess,
		uintptr(0x0008), // TOKEN_QUERY
		uintptr(unsafe.Pointer(&token)),
	)
	if ret == 0 {
		return false
	}
	defer syscall.CloseHandle(token)

	// 查询令牌信息
	var elevation uint32
	var returnedLen uint32
	ret, _, _ = getTokenInformation.Call(
		uintptr(token),
		uintptr(20), // TokenElevation
		uintptr(unsafe.Pointer(&elevation)),
		uintptr(unsafe.Sizeof(elevation)),
		uintptr(unsafe.Pointer(&returnedLen)),
	)
	if ret == 0 {
		return false
	}

	llog.Info("当前程序以管理员身份运行: %v", elevation != 0)

	return elevation != 0
}

// BringWindowToFront 将指定标题的窗口置顶到前台
// windowTitle: 要搜索的窗口标题（支持部分匹配）
// 返回值: 成功返回true，失败返回false和错误信息
func BringWindowToFront(windowTitle string) (bool, error) {
	if windowTitle == "" {
		return false, fmt.Errorf("窗口标题不能为空")
	}

	var targetHwnd uintptr
	var foundTitle string

	// 枚举所有窗口的回调函数
	cb := syscall.NewCallback(func(hwnd uintptr, lparam uintptr) uintptr {
		buf := make([]uint16, 512) // 增加缓冲区大小
		ret, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))

		if ret == 0 {
			return 1 // 继续枚举
		}

		title := syscall.UTF16ToString(buf)
		if title != "" && strings.Contains(title, windowTitle) {
			visible, _, _ := procIsWindowVisible.Call(hwnd)
			if visible != 0 {
				foundTitle = title
				targetHwnd = hwnd
				return 0 // 停止枚举
			}
		}
		return 1 // 继续枚举
	})

	// 枚举所有窗口
	ret, _, _ := procEnumWindows.Call(cb, 0)
	if ret == 0 && targetHwnd == 0 {
		return false, fmt.Errorf("枚举窗口失败")
	}

	// 先尝试恢复窗口（如果是最小化状态）
	ret, _, _ = procShowWindow.Call(targetHwnd, SW_RESTORE)
	if ret == 0 {
		// 如果恢复失败，尝试显示窗口
		procShowWindow.Call(targetHwnd, SW_SHOW)
	}

	// 设置窗口到前台
	ret, _, err := procSetForegroundWindow.Call(targetHwnd)
	if ret == 0 {
		return false, fmt.Errorf("设置窗口到前台失败: %v (窗口: %s)", err, foundTitle)
	}

	llog.Debug("窗口 '%s' 已经被置顶", foundTitle)

	return true, nil
}
