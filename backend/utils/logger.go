package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Logger 日志记录器
type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

// NewLogger 创建新的日志记录器
func NewLogger() *Logger {
	// 确保日志目录存在
	logDir := "./logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("创建日志目录失败: %v", err)
	}

	// 创建日志文件
	logFile := filepath.Join(logDir, time.Now().Format("2006-01-02")+".log")
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("打开日志文件失败: %v", err)
	}

	// 创建日志记录器
	infoLogger := log.New(file, "INFO: ", log.Ldate|log.Ltime)
	errorLogger := log.New(file, "ERROR: ", log.Ldate|log.Ltime)

	return &Logger{
		infoLogger:  infoLogger,
		errorLogger: errorLogger,
	}
}

// getCallerInfo 获取调用者信息
func getCallerInfo() string {
	_, file, line, ok := runtime.Caller(2) // 跳过两层调用栈
	if !ok {
		return "unknown:0"
	}
	// 提取文件名
	fileName := filepath.Base(file)
	return fmt.Sprintf("%s:%d", fileName, line)
}

// Info 记录信息日志
func (l *Logger) Info(format string, v ...interface{}) {
	callerInfo := getCallerInfo()
	message := fmt.Sprintf(format, v...)
	l.infoLogger.Printf("[%s] %s", callerInfo, message)
}

// Error 记录错误日志
func (l *Logger) Error(format string, v ...interface{}) {
	callerInfo := getCallerInfo()
	message := fmt.Sprintf(format, v...)
	l.errorLogger.Printf("[%s] %s", callerInfo, message)
}

// 全局日志记录器
var globalLogger *Logger

// 初始化全局日志记录器
func init() {
	globalLogger = NewLogger()
}

// Info 全局信息日志
func Info(format string, v ...interface{}) {
	globalLogger.Info(format, v...)
}

// Error 全局错误日志
func Error(format string, v ...interface{}) {
	globalLogger.Error(format, v...)
}
