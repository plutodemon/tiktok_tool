package lkit

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"
)

// SigChan 创建一个通道来接收信号
var SigChan = make(chan os.Signal)

func AnyToStr(v interface{}) string {
	return fmt.Sprint(v)
}

func SliceToStrList[k comparable](v []k) []string {
	var res []string
	for _, i := range v {
		res = append(res, fmt.Sprint(i))
	}
	return res
}

func GetAddr(host, port any) string {
	return fmt.Sprintf("%s:%d", host, port)
}

// GetNowUnix 秒级时间戳
func GetNowUnix() int64 {
	return time.Now().Unix()
}

// GetPathDir 获得全路径下的目录
func GetPathDir(path string) string {
	return filepath.Dir(filepath.Clean(filepath.FromSlash(path)))
}

var IsAdmin bool

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

	return elevation != 0
}
