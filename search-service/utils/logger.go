package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// ============================================================
// 日志工具
// 输出 JSON 行格式，兼容 Filebeat 等日志采集工具解析
// 支持异步写入、缓冲通道、日志文件轮转
// ============================================================

// logEntry 日志条目结构体，序列化为 JSON 后输出
type logEntry struct {
	// Timestamp 日志产生时间，ISO8601 格式
	Timestamp string `json:"timestamp"`

	// Level 日志级别：INFO / WARN / ERROR
	Level string `json:"level"`

	// Message 日志内容
	Message string `json:"message"`

	// Caller 调用位置，格式：文件名:行号
	Caller string `json:"caller"`

	// Service 服务名称，用于区分不同微服务的日志
	Service string `json:"service"`
}

// 日志级别常量
const (
	levelInfo  = "INFO"
	levelWarn  = "WARN"
	levelError = "ERROR"
)

// 全局变量
var (
	// logChannel 异步日志缓冲通道，容量 1000 条
	logChannel chan logEntry

	// serviceName 当前服务名称，在 processLogs 中设置
	serviceName string

	// logFile 当前日志文件句柄
	logFile *os.File

	// currentLogSize 当前日志文件大小（字节）
	currentLogSize int64

	// maxLogSize 日志文件最大大小，超过此值进行轮转（10MB）
	maxLogSize int64 = 10 * 1024 * 1024

	// logFilePath 日志文件路径
	logFilePath string

	// once 确保 logger 只初始化一次
	once sync.Once

	// wg 用于等待日志处理协程退出
	wg sync.WaitGroup
)

// InitLogger 初始化日志系统
// name: 服务名称，会写入每条日志的 service 字段
// 创建日志缓冲通道，启动后台协程处理日志写入
func InitLogger(name string) {
	once.Do(func() {
		serviceName = name
		logChannel = make(chan logEntry, 1000)

		// 创建 logs 目录
		logDir := "logs"
		if err := os.MkdirAll(logDir, 0755); err != nil {
			// 如果无法创建日志目录，仅输出到 stdout
			fmt.Fprintf(os.Stderr, "无法创建日志目录: %v\n", err)
		}

		// 打开日志文件
		logFilePath = filepath.Join(logDir, name+".log")
		openLogFile()

		// 启动后台协程处理日志写入
		wg.Add(1)
		go processLogs()
	})
}

// openLogFile 打开或创建日志文件
// 如果文件已存在，获取当前文件大小用于轮转判断
func openLogFile() {
	f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法打开日志文件: %v\n", err)
		return
	}

	// 获取当前文件大小
	info, err := f.Stat()
	if err != nil {
		currentLogSize = 0
	} else {
		currentLogSize = info.Size()
	}

	logFile = f
}

// rotateLogFile 日志文件轮转
// 当日志文件超过 maxLogSize 时，将当前文件重命名为 .old，然后创建新文件
func rotateLogFile() {
	if logFile != nil {
		logFile.Close()
	}

	// 将当前日志文件重命名为 .old
	oldPath := logFilePath + ".old"
	os.Rename(logFilePath, oldPath)

	// 创建新的日志文件
	openLogFile()
}

// processLogs 后台协程，从缓冲通道读取日志条目并写入文件和标准输出
// 同时负责日志文件轮转检查
func processLogs() {
	defer wg.Done()

	for entry := range logChannel {
		// 设置服务名称
		entry.Service = serviceName

		// 序列化为 JSON
		data, err := json.Marshal(entry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "日志序列化失败: %v\n", err)
			continue
		}

		// 写入日志文件
		if logFile != nil {
			line := string(data) + "\n"
			n, err := logFile.WriteString(line)
			if err != nil {
				fmt.Fprintf(os.Stderr, "日志写入文件失败: %v\n", err)
			} else {
				currentLogSize += int64(n)
				// 检查是否需要轮转
				if currentLogSize >= maxLogSize {
					rotateLogFile()
				}
			}
		}

		// 同时输出到标准输出（方便开发调试）
		fmt.Println(string(data))
	}
}

// getCaller 获取调用者的文件名和行号
// skip: 跳过的调用栈层数，0 表示 getCaller 自身
func getCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip + 2) // +2 跳过 getCaller 和 Info/Warn/Error
	if !ok {
		return "unknown"
	}
	// 只取文件名，不取完整路径
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

// Info 记录信息级别日志
// 用于记录正常运行状态的信息，如服务启动、请求处理完成等
// format: 格式化字符串，args: 格式化参数
func Info(format string, args ...interface{}) {
	if logChannel == nil {
		return
	}
	logChannel <- logEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     levelInfo,
		Message:   fmt.Sprintf(format, args...),
		Caller:    getCaller(1),
	}
}

// Warn 记录警告级别日志
// 用于记录不影响运行但需要关注的情况，如接近限流阈值、配置项缺失使用默认值等
// format: 格式化字符串，args: 格式化参数
func Warn(format string, args ...interface{}) {
	if logChannel == nil {
		return
	}
	logChannel <- logEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     levelWarn,
		Message:   fmt.Sprintf(format, args...),
		Caller:    getCaller(1),
	}
}

// Error 记录错误级别日志
// 用于记录影响功能的错误，如数据库连接失败、ES 查询异常等
// format: 格式化字符串，args: 格式化参数
func Error(format string, args ...interface{}) {
	if logChannel == nil {
		return
	}
	logChannel <- logEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     levelError,
		Message:   fmt.Sprintf(format, args...),
		Caller:    getCaller(1),
	}
}

// CloseLogger 关闭日志系统
// 关闭缓冲通道，等待后台协程处理完剩余日志后退出
// 应在程序退出前调用，确保所有日志都被写入
func CloseLogger() {
	if logChannel != nil {
		close(logChannel)
		wg.Wait()
	}
	if logFile != nil {
		logFile.Close()
	}
}
