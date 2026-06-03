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
	currentFile *os.File
	logDir      string
}

// logEntry 日志条目
type logEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Caller    string `json:"caller"`
	Service   string `json:"service"`
}

// MaxLogFileSize 日志文件最大大小 (10MB)
const MaxLogFileSize = 10 * 1024 * 1024

// NewLogger 创建新的日志记录器
func NewLogger() *Logger {
	logDir := "./logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("创建日志目录失败: %v", err)
	}

	logFile, err := os.OpenFile(generateLogFilePath(logDir), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("打开日志文件失败: %v", err)
	}

	infoLogger := log.New(logFile, "", 0)
	warnLogger := log.New(logFile, "", 0)
	errorLogger := log.New(logFile, "", 0)

	logger := &Logger{
		infoLogger:  infoLogger,
		warnLogger:  warnLogger,
		errorLogger: errorLogger,
		logChan:     make(chan logEntry, 1000),
		currentFile: logFile,
		logDir:      logDir,
	}

	logger.wg.Add(1)
	go logger.processLogs("shop-backend")

	return logger
}

func generateLogFilePath(logDir string) string {
	baseLogName := time.Now().Format("2006-01-02")
	ext := ".log"

	for i := 1; ; i++ {
		filePath := filepath.Join(logDir, fmt.Sprintf("%s_%d%s", baseLogName, i, ext))
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			if i == 1 {
				basePath := filepath.Join(logDir, baseLogName+ext)
				if _, err := os.Stat(basePath); os.IsNotExist(err) {
					return basePath
				}
			}
			return filePath
		}
	}
}

func (l *Logger) checkAndRotate() {
	if l.currentFile == nil {
		return
	}

	info, err := l.currentFile.Stat()
	if err != nil {
		return
	}

	if info.Size() >= MaxLogFileSize {
		l.rotateLogFile()
	}
}

func (l *Logger) rotateLogFile() {
	if l.currentFile != nil {
		l.currentFile.Close()
	}

	newFilePath := l.generateNewLogFilePath()
	file, err := os.OpenFile(newFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("创建新日志文件失败: %v", err)
		return
	}

	l.currentFile = file
	l.infoLogger.SetOutput(file)
	l.warnLogger.SetOutput(file)
	l.errorLogger.SetOutput(file)
}

func (l *Logger) generateNewLogFilePath() string {
	baseLogName := time.Now().Format("2006-01-02")
	ext := ".log"

	for i := 1; ; i++ {
		filePath := filepath.Join(l.logDir, fmt.Sprintf("%s_%d%s", baseLogName, i, ext))
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return filePath
		}
	}
}

// processLogs 处理日志队列
func (l *Logger) processLogs(serviceName string) {
	defer l.wg.Done()
	for entry := range l.logChan {
		l.checkAndRotate()

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
	if l.currentFile != nil {
		l.currentFile.Close()
	}
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
