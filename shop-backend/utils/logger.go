package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Logger 日志记录器
type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	logChan     chan logEntry
	wg          sync.WaitGroup
}

// logEntry 日志条目
type logEntry struct {
	level  string
	format string
	args   []interface{}
	caller string
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

	logger := &Logger{
		infoLogger:  infoLogger,
		errorLogger: errorLogger,
		logChan:     make(chan logEntry, 1000), // 带缓冲的通道
	}

	// 启动日志处理协程
	logger.wg.Add(1)
	go logger.processLogs()

	return logger
}

// processLogs 处理日志队列
func (l *Logger) processLogs() {
	defer l.wg.Done()
	for entry := range l.logChan {
		callerInfo := entry.caller
		message := fmt.Sprintf(entry.format, entry.args...)
		switch entry.level {
		case "info":
			l.infoLogger.Printf("[%s] %s", callerInfo, message)
		case "error":
			l.errorLogger.Printf("[%s] %s", callerInfo, message)
		}
	}
}

// getCallerInfo 获取调用者信息
func getCallerInfo() string {
	_, file, line, ok := runtime.Caller(3) // 跳过三层调用栈
	if !ok {
		return "unknown:0"
	}
	// 提取文件名
	fileName := filepath.Base(file)
	return fmt.Sprintf("%s:%d", fileName, line)
}

// Info 记录信息日志
func (l *Logger) Info(format string, v ...interface{}) {
	l.logChan <- logEntry{
		level:  "info",
		format: format,
		args:   v,
		caller: getCallerInfo(),
	}
}

// Error 记录错误日志
func (l *Logger) Error(format string, v ...interface{}) {
	l.logChan <- logEntry{
		level:  "error",
		format: format,
		args:   v,
		caller: getCallerInfo(),
	}
}

// Close 关闭日志记录器，确保所有日志都被处理
func (l *Logger) Close() {
	close(l.logChan)
	l.wg.Wait()
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

// CloseLogger 关闭全局日志记录器
func CloseLogger() {
	globalLogger.Close()
}
