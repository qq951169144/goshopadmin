package utils

import (
	"encoding/json"
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
	warnLogger  *log.Logger
	errorLogger *log.Logger
	logChan     chan logEntry
	wg          sync.WaitGroup
}

// logEntry 日志条目
type logEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Caller    string `json:"caller"`
	Service   string `json:"service"`
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

	// 创建日志记录器 - 不使用前缀，我们自己格式化
	infoLogger := log.New(file, "", 0)
	warnLogger := log.New(file, "", 0)
	errorLogger := log.New(file, "", 0)

	logger := &Logger{
		infoLogger:  infoLogger,
		warnLogger:  warnLogger,
		errorLogger: errorLogger,
		logChan:     make(chan logEntry, 1000), // 带缓冲的通道
	}

	// 启动日志处理协程
	logger.wg.Add(1)
	go logger.processLogs("backend")

	return logger
}

// processLogs 处理日志队列
func (l *Logger) processLogs(serviceName string) {
	defer l.wg.Done()
	for entry := range l.logChan {
		entry.Service = serviceName
		entry.Timestamp = time.Now().Format("2006-01-02T15:04:05Z07:00")

		jsonData, err := json.Marshal(entry)
		if err != nil {
			log.Printf("日志序列化失败: %v", err)
			continue
		}

		switch entry.Level {
		case "INFO":
			l.infoLogger.Println(string(jsonData))
		case "WARN":
			l.warnLogger.Println(string(jsonData))
		case "ERROR":
			l.errorLogger.Println(string(jsonData))
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
		Level:   "INFO",
		Message: fmt.Sprintf(format, v...),
		Caller:  getCallerInfo(),
	}
}

// Warn 记录警告日志
func (l *Logger) Warn(format string, v ...interface{}) {
	l.logChan <- logEntry{
		Level:   "WARN",
		Message: fmt.Sprintf(format, v...),
		Caller:  getCallerInfo(),
	}
}

// Error 记录错误日志
func (l *Logger) Error(format string, v ...interface{}) {
	l.logChan <- logEntry{
		Level:   "ERROR",
		Message: fmt.Sprintf(format, v...),
		Caller:  getCallerInfo(),
	}
}

// Close 关闭日志记录器，确保所有日志都被处理
func (l *Logger) Close() {
	close(l.logChan)
	l.wg.Wait()
}

// 全局日志记录器
var (
	globalLogger *Logger
	loggerOnce   sync.Once
)

// InitLogger 初始化全局日志记录器，由 main() 显式调用
func InitLogger() {
	loggerOnce.Do(func() {
		globalLogger = NewLogger()
	})
}

// Info 全局信息日志
func Info(format string, v ...interface{}) {
	if globalLogger == nil {
		return
	}
	globalLogger.Info(format, v...)
}

// Warn 全局警告日志
func Warn(format string, v ...interface{}) {
	if globalLogger == nil {
		return
	}
	globalLogger.Warn(format, v...)
}

// Error 全局错误日志
func Error(format string, v ...interface{}) {
	if globalLogger == nil {
		return
	}
	globalLogger.Error(format, v...)
}

// CloseLogger 关闭全局日志记录器
func CloseLogger() {
	if globalLogger == nil {
		return
	}
	globalLogger.Close()
}
