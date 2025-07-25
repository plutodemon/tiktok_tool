package lkit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"tiktok_tool/config"
	"tiktok_tool/llog"
	"time"
	"unsafe"

	"github.com/shirou/gopsutil/v4/process"
	"golang.org/x/sys/windows"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const MaxUInt32 = 0xffffffff         // DEC 4294967295
const MaxInt32 = 0x7fffffff          // DEC 2147483647
const MaxInt64 = 0x7fffffffffffffff  // 9223372036854775807
const MaxUInt64 = 0xffffffffffffffff // 18446744073709551615
const MaxInt64Half = 0x7fffffffffffffff / 2

// SigChan 创建一个通道来接收信号
var SigChan = make(chan os.Signal)

func AnyToStr(v interface{}) string {
	return fmt.Sprint(v)
}

func Str2Int32(str string) int32 {
	num := Str2Int64(str)
	if num > MaxInt32 {
		llog.Error("Str2Int32 overflow", str)
	}
	return int32(num)
}
func TryParseStr2Int64(str string) (int64, error) {
	i64, err := strconv.ParseInt(str, 10, 0)
	if err != nil {
		return 0, err
	}
	return i64, nil
}

func Str2Int64(str string) int64 {
	i64, err := TryParseStr2Int64(str)
	if err != nil {
		llog.Error(err.Error(), 0, "[Str2Int64]", "["+str+"]")
		return 0
	}
	return i64
}
func Str2UInt64(str string) uint64 {
	return uint64(Str2Int64(str))
}

func Str2UInt32(str string) uint32 {
	num := Str2Int64(str)
	if num > MaxUInt32 {
		llog.Error("Str2Int32 overflow", str)
	}
	return uint32(num)
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
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

// RunAutoTool 执行外部工具程序并解析JSON输出
// exePath: 可执行文件路径
// args: 命令行参数
// 返回值: 解析后的AutoResult结构体和可能的错误
func RunAutoTool(exePath string, args []string) (*AutoResult, error) {
	args = append(args,
		"--check-interval", AnyToStr(config.GetConfig().PluginCheckInterval),
		"--wait-after-found", AnyToStr(config.GetConfig().PluginWaitAfterFound),
		"--timeout", AnyToStr(config.GetConfig().PluginTimeout),
	)
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

	// 鼠标输入相关常量
	INPUT_MOUSE           = 0
	MOUSEEVENTF_LEFTDOWN  = 0x0002
	MOUSEEVENTF_LEFTUP    = 0x0004
	MOUSEEVENTF_RIGHTDOWN = 0x0008
	MOUSEEVENTF_RIGHTUP   = 0x0010
	MOUSEEVENTF_ABSOLUTE  = 0x8000
)

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

	return true, nil
}

// INPUT 结构体用于SendInput API
type INPUT struct {
	Type uint32
	Mi   MouseInput
}

// MouseInput 结构体定义鼠标输入
type MouseInput struct {
	Dx          int32
	Dy          int32
	MouseData   uint32
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

// SimulateMouseClick 模拟鼠标点击指定坐标
// x: 屏幕X坐标
// y: 屏幕Y坐标
// button: 鼠标按钮类型 ("left" 或 "right")
// 返回值: 成功返回nil，失败返回错误信息
func SimulateMouseClick(x, y int, button string) error {
	if x < 0 || y < 0 {
		return fmt.Errorf("坐标不能为负数: x=%d, y=%d", x, y)
	}

	// 移动鼠标到指定位置
	ret, _, err := procSetCursorPos.Call(uintptr(x), uintptr(y))
	if ret == 0 {
		return fmt.Errorf("移动鼠标失败: %v", err)
	}

	// 短暂延迟确保鼠标移动完成
	time.Sleep(10 * time.Millisecond)

	// 根据按钮类型设置鼠标事件标志
	var downFlag, upFlag uint32
	switch strings.ToLower(button) {
	case "left", "":
		downFlag = MOUSEEVENTF_LEFTDOWN
		upFlag = MOUSEEVENTF_LEFTUP
	case "right":
		downFlag = MOUSEEVENTF_RIGHTDOWN
		upFlag = MOUSEEVENTF_RIGHTUP
	default:
		return fmt.Errorf("不支持的鼠标按钮类型: %s (支持 'left' 或 'right')", button)
	}

	// 创建鼠标按下事件
	inputDown := INPUT{
		Type: INPUT_MOUSE,
		Mi: MouseInput{
			Dx:      0,
			Dy:      0,
			DwFlags: downFlag,
		},
	}

	// 创建鼠标释放事件
	inputUp := INPUT{
		Type: INPUT_MOUSE,
		Mi: MouseInput{
			Dx:      0,
			Dy:      0,
			DwFlags: upFlag,
		},
	}

	// 发送鼠标按下事件
	ret, _, err = procSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&inputDown)),
		uintptr(unsafe.Sizeof(inputDown)),
	)
	if ret == 0 {
		return fmt.Errorf("发送鼠标按下事件失败: %v", err)
	}

	// 短暂延迟模拟真实点击
	time.Sleep(50 * time.Millisecond)

	// 发送鼠标释放事件
	ret, _, err = procSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&inputUp)),
		uintptr(unsafe.Sizeof(inputUp)),
	)
	if ret == 0 {
		return fmt.Errorf("发送鼠标释放事件失败: %v", err)
	}

	return nil
}

// SimulateLeftClick 模拟鼠标左键点击指定坐标的便捷方法
// x: 屏幕X坐标
// y: 屏幕Y坐标
// 返回值: 成功返回nil，失败返回错误信息
func SimulateLeftClick(x, y int) error {
	return SimulateMouseClick(x, y, "left")
}

// SimulateRightClick 模拟鼠标右键点击指定坐标的便捷方法
// x: 屏幕X坐标
// y: 屏幕Y坐标
// 返回值: 成功返回nil，失败返回错误信息
func SimulateRightClick(x, y int) error {
	return SimulateMouseClick(x, y, "right")
}
