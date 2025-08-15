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
	"time"

	"github.com/shirou/gopsutil/v4/process"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"tiktok_tool/config"
	"tiktok_tool/llog"
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

	llog.Info("进程已终止", pid)

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
	llog.Debug("运行自动化工具:", args)
	cfg := config.GetConfig().ScriptSettings
	args = append(args,
		"--check-interval", AnyToStr(cfg.PluginCheckInterval),
		"--wait-after-found", AnyToStr(cfg.PluginWaitAfterFound),
		"--timeout", AnyToStr(cfg.PluginTimeout),
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
