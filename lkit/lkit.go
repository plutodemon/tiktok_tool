package lkit

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
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
