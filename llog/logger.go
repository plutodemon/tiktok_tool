package llog

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	log   *zap.Logger
	sugar *zap.SugaredLogger
	once  sync.Once
)

// 初始化一个空的logger，避免在未初始化前调用日志函数导致空指针异常
func init() {
	// 创建一个nopLogger，所有操作都是空操作
	nopCore := zapcore.NewNopCore()
	log = zap.New(nopCore)
	sugar = log.Sugar()
}

// Init 初始化日志系统
func Init(logConfig *LogSetting) error {
	var initErr error

	// 使用sync.Once确保只初始化一次
	once.Do(func() {
		initErr = initLogger(logConfig)
	})

	return initErr
}

// initLogger 实际的日志初始化函数
func initLogger(logConfig *LogSetting) error {
	if logConfig == nil {
		logConfig = &DefaultConfig
	}

	// 创建日志目录
	if logConfig.File {
		logDir := filepath.Clean(logConfig.FilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("创建日志目录失败: %v", err)
		}
	}

	// 设置日志级别
	logLevel := getLogLevel(logConfig.Level)

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:       "msg",
		LevelKey:         "level",
		TimeKey:          "time",
		NameKey:          "logger",
		CallerKey:        "caller",
		FunctionKey:      zapcore.OmitKey,
		StacktraceKey:    "stacktrace",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		EncodeTime:       zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: "\t",
	}

	var cores []zapcore.Core

	// 添加控制台输出
	if logConfig.Console {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), logLevel))
	}

	// 添加文件输出
	if logConfig.File {
		// 使用当前日期作为日志文件名
		fileName := fmt.Sprintf(logConfig.Format, time.Now().Format("2006-01-02"))
		path := filepath.Join(logConfig.FilePath, fileName)

		var fileEncoder zapcore.Encoder
		if logConfig.OutputFormat != "text" {
			fileEncoder = zapcore.NewJSONEncoder(encoderConfig)
		} else {
			fileEncoder = zapcore.NewConsoleEncoder(encoderConfig)
		}

		writer := zapcore.AddSync(&lumberjack.Logger{
			Filename:   filepath.Clean(path),
			MaxSize:    logConfig.MaxSize,
			MaxBackups: logConfig.MaxBackups,
			MaxAge:     logConfig.MaxAge,
			Compress:   logConfig.Compress,
			LocalTime:  logConfig.LocalTime,
		})
		cores = append(cores, zapcore.NewCore(fileEncoder, writer, logLevel))
	}

	// 创建核心
	core := zapcore.NewTee(cores...)

	// 创建logger
	log = zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	// 创建sugar logger
	sugar = log.Sugar()

	_, err := zap.RedirectStdLogAt(log, zapcore.ErrorLevel)
	if err != nil {
		return fmt.Errorf("重定向标准日志失败: %v", zap.Error(err))
	}

	// 输出初始化信息
	sugar.Info("日志系统初始化完成，日志级别:", strings.ToUpper(logConfig.Level))

	return nil
}

// FormatError 格式化错误信息，去除重复
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	parts := strings.Split(err.Error(), ": ")
	seen := make(map[string]bool)
	var unique []string

	for _, part := range parts {
		if !seen[part] {
			seen[part] = true
			unique = append(unique, part)
		}
	}

	return strings.Join(unique, ": ")
}

// Debug 输出调试级别日志
func Debug(args ...interface{}) {
	sugar.Debug(args...)
}

// Info 输出信息级别日志
func Info(args ...interface{}) {
	sugar.Info(args...)
}

// Warn 输出警告级别日志
func Warn(args ...interface{}) {
	sugar.Warn(args...)
}

// Error 输出错误级别日志
func Error(args ...interface{}) {
	sugar.Error(args...)
}

// Fatal 输出致命错误日志并退出程序
func Fatal(args ...interface{}) {
	sugar.Fatal(args...)
}

func DebugF(format string, args ...interface{}) {
	sugar.Debugf(format, args...)
}

func InfoF(format string, args ...interface{}) {
	sugar.Infof(format, args...)
}

func WarnF(format string, args ...interface{}) {
	sugar.Warnf(format, args...)
}

func ErrorF(format string, args ...interface{}) {
	sugar.Errorf(format, args...)
}

func FatalF(format string, args ...interface{}) {
	sugar.Fatalf(format, args...)
}

// Sync 同步日志到磁盘
func Sync() {
	if log != nil {
		return
	}

	_ = log.Sync()
	if sugar != nil {
		_ = sugar.Sync()
	}
}

// Cleanup 清理日志资源
func Cleanup() {
	if log != nil {
		Sync()
	}

	// 重置为nopLogger
	nopCore := zapcore.NewNopCore()
	log = zap.New(nopCore)
	sugar = log.Sugar()
}

// HandlePanic 处理panic并记录日志
func HandlePanic() {
	if r := recover(); r != nil {
		Fatal(fmt.Sprintf("程序发生严重错误: %v\n堆栈信息:\n%s", r, debug.Stack()))
	}
}
