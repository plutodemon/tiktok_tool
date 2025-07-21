package lkit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/shirou/gopsutil/v4/process"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
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

// IsProcessRunning 检查指定进程是否正在运行
func IsProcessRunning(targetName ...string) ([]int32, error) {
	processes, err := process.Processes()
	if err != nil {
		return []int32{}, err
	}

	ret := make([]int32, len(targetName))
	for _, p := range processes {
		name, err := p.Name()
		if err != nil {
			continue
		}
		for index, nameInfo := range targetName {
			if strings.EqualFold(name, nameInfo) {
				ret[index] = p.Pid
			}
		}
	}

	return ret, nil
}

func KillProcess(pid int32) error {
	p, err := process.NewProcess(pid)
	if err != nil {
		return fmt.Errorf("无法获取进程信息: %w", err)
	}

	if err = p.Kill(); err != nil {
		return fmt.Errorf("无法终止进程: %w", err)
	}

	return nil
}

// gbkToUTF8 将GBK编码的字节数组转换为UTF-8字符串
// data: GBK编码的字节数组
// 返回值: UTF-8字符串和可能的错误
func gbkToUTF8(data []byte) (string, error) {
	reader := transform.NewReader(strings.NewReader(string(data)), simplifiedchinese.GBK.NewDecoder())
	result, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

type AutoResult struct {
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
	ControlTitle string `json:"control_title,omitempty"`
	ControlType  string `json:"control_type,omitempty"`
	Position     struct {
		Left   int `json:"left"`
		Top    int `json:"top"`
		Right  int `json:"right"`
		Bottom int `json:"bottom"`
	} `json:"position,omitempty"`
	Center struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"center,omitempty"`
	Clicked      bool   `json:"clicked,omitempty"`
	Action       string `json:"action,omitempty"`
	AppName      string `json:"app_name,omitempty"`
	WindowsFound int    `json:"windows_found,omitempty"`
	DumpFile     string `json:"dump_file,omitempty"`
}

func RunAutoTool(exePath string, args []string) (*AutoResult, error) {
	cmd := exec.Command(exePath, args...)

	// 设置工作目录
	cmd.Dir = filepath.Dir(exePath)

	// 执行命令并获取输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("执行命令失败: %v, 输出: %s", err, string(output))
	}

	// 解析JSON输出
	var result AutoResult

	err = json.Unmarshal(output, &result)
	if err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v, 原始输出: %s", err, output)
	}

	return &result, nil
}
