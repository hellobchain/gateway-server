package logger

import (
	"io"
	"strings"

	"github.com/hellobchain/wswlog/wlogging"
)

var logger = wlogging.MustGetFileLoggerWithoutName(LogConfig)

// 设置程序日志级别
func SetLogLevel(logLevel string) {
	logLevel = strings.ToLower(logLevel)
	logger.Debug("日志级别", logLevel)
	wlogging.SetGlobalLogLevel(logLevel)
}

// 设置是否输出到控制台
func SetConsole(isConsole bool) {
	// 改进日志内容，避免直接输出布尔值
	if isConsole {
		logger.Debug("启用控制台输出")
	} else {
		logger.Debug("禁用控制台输出")
	}
	wlogging.SetConsole(isConsole)
}

// 设置默认的日志输出
func SetDefaultWriter(w io.Writer) {
	// 检查传入的 writer 是否为 nil，避免潜在的运行时 panic
	if w == nil {
		logger.Error("SetDefaultWriter: writer cannot be nil")
		return
	}
	wlogging.SetDefaultWriter(w)
}

var LogConfig = &wlogging.LogConfig{
	LogPath:      "logs/system.log",
	MaxAge:       DEFAULT_MAX_AGE,
	RotationTime: DEFAULT_ROTATION_TIME,
}

// 日志切割默认配置
const (
	DEFAULT_MAX_AGE       = 7   // 日志最长保存时间，单位：天
	DEFAULT_ROTATION_TIME = 24  // 日志滚动间隔，单位：小时
	DEFAULT_ROTATION_SIZE = 100 // 默认的日志滚动大小，单位：MB
)
