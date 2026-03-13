package utils

import (
	"bytes"
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
	go logger.processLogs()

	return logger
}

// processLogs 处理日志队列
func (l *Logger) processLogs() {
	defer l.wg.Done()
	for entry := range l.logChan {
		message := fmt.Sprintf(entry.format, entry.args...)

		// 尝试格式化 JSON
		formattedMsg := formatLogMessage(entry.level, entry.caller, message)

		switch entry.level {
		case "INFO":
			l.infoLogger.Println(formattedMsg)
		case "WARN":
			l.warnLogger.Println(formattedMsg)
		case "ERROR":
			l.errorLogger.Println(formattedMsg)
		}
	}
}

// formatLogMessage 格式化日志消息，如果是 JSON 则格式化输出
func formatLogMessage(level, caller, message string) string {
	timestamp := time.Now().Format("2006/01/02 15:04:05")

	// 尝试解析 message 是否为 JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(message), &jsonData); err == nil {
		// 是 JSON，格式化输出
		var buf bytes.Buffer
		encoder := json.NewEncoder(&buf)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(jsonData); err == nil {
			// 移除最后的换行符
			formattedJSON := buf.String()
			if len(formattedJSON) > 0 && formattedJSON[len(formattedJSON)-1] == '\n' {
				formattedJSON = formattedJSON[:len(formattedJSON)-1]
			}
			return fmt.Sprintf("%s [%s] [%s]\n%s", timestamp, level, caller, formattedJSON)
		}
	}

	// 不是 JSON，普通格式输出
	return fmt.Sprintf("%s [%s] [%s] %s", timestamp, level, caller, message)
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
		level:  "INFO",
		format: format,
		args:   v,
		caller: getCallerInfo(),
	}
}

// Warn 记录警告日志
func (l *Logger) Warn(format string, v ...interface{}) {
	l.logChan <- logEntry{
		level:  "WARN",
		format: format,
		args:   v,
		caller: getCallerInfo(),
	}
}

// Error 记录错误日志
func (l *Logger) Error(format string, v ...interface{}) {
	l.logChan <- logEntry{
		level:  "ERROR",
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

// Warn 全局警告日志
func Warn(format string, v ...interface{}) {
	globalLogger.Warn(format, v...)
}

// Error 全局错误日志
func Error(format string, v ...interface{}) {
	globalLogger.Error(format, v...)
}

// CloseLogger 关闭全局日志记录器
func CloseLogger() {
	globalLogger.Close()
}
