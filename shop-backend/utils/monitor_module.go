package utils

import (
	"bytes"
	"runtime"
	"strings"
)

// collectModuleStats 采集按模块分组的协程统计
// 通过 runtime.Stack() 获取所有协程的堆栈信息，解析出每个协程所属的模块
func (m *Monitor) collectModuleStats() ModuleStats {
	stats := make(ModuleStats)

	// runtime.Stack(buf, true) 将所有活跃协程的堆栈写入 buf
	// 参数 true 表示获取所有协程（而非仅当前协程）
	// 返回值 n 为实际写入的字节数
	// 复用 Monitor.stackBuf 避免每次采集分配 1MB 内存
	n := runtime.Stack(m.stackBuf, true)
	stackData := m.stackBuf[:n]

	// 协程堆栈以 "\n\n" 分隔，每个块代表一个协程的完整堆栈
	sep := []byte("\n\n")
	goroutines := bytes.Split(stackData, sep)

	for _, goroutine := range goroutines {
		module := extractModule(goroutine)
		stats[module]++
	}

	return stats
}

// extractModule 从单个协程的堆栈数据中提取所属模块名
// 遍历堆栈的每一行，找到第一个匹配已知模块前缀的行即返回
// 如果没有匹配任何已知模块，返回 "other"
func extractModule(goroutineStack []byte) string {
	lines := bytes.Split(goroutineStack, []byte("\n"))
	// 从索引 1 开始，跳过第一行的 goroutine 头信息（如 "goroutine 123 [running]")
	// i += 2 是因为堆栈格式：函数调用行 + 参数/源代码行
	for i := 1; i < len(lines); i += 2 {
		line := strings.TrimSpace(string(lines[i]))
		module := parseModuleFromStackLine(line)
		if module != "" {
			return module
		}
	}
	return "other"
}

// parseModuleFromStackLine 从堆栈单行中解析模块名
// 堆栈行格式示例：shop-backend/services.(*OrderService).CreateOrder(...)
// 通过匹配已知模块路径前缀，提取模块名（如 services、controllers 等）
func parseModuleFromStackLine(line string) string {
	knownModules := []string{
		"shop-backend/services",
		"shop-backend/controllers",
		"shop-backend/mq",
		"shop-backend/utils",
		"shop-backend/middleware",
		"shop-backend/cache",
		"shop-backend/config",
		"shop-backend/routes",
		"shop-backend/pkg",
	}

	for _, mod := range knownModules {
		if strings.Contains(line, mod) {
			// 从完整路径中提取最后一段作为模块名
			// 例如 "shop-backend/services" → "services"
			parts := strings.Split(mod, "/")
			return parts[len(parts)-1]
		}
	}

	// Go 运行时内部协程（如 GC 协程、调度器协程等）
	if strings.Contains(line, "runtime.") {
		return "runtime"
	}

	return ""
}
