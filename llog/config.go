package llog

import (
	"strings"

	"go.uber.org/zap/zapcore"
)

// 添加日志级别映射缓存
var levelMap = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
	"fatal": zapcore.FatalLevel,
}

// getLogLevel 获取日志级别
func getLogLevel(level string) zapcore.Level {
	if level == "" {
		return zapcore.InfoLevel
	}
	if l, ok := levelMap[strings.ToLower(level)]; ok {
		return l
	}
	return zapcore.InfoLevel
}

// DefaultConfig 默认日志配置
var DefaultConfig = LogSetting{
	Console:      true,
	File:         true,
	FilePath:     "logs",
	MaxSize:      64,
	MaxAge:       7,
	MaxBackups:   30,
	Compress:     true,
	LocalTime:    true,
	Format:       "%s.log",
	Level:        "debug",
	OutputFormat: "text",
}
