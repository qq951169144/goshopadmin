package utils

import (
	"bytes"
	"runtime"
	"strings"
)

func collectModuleStats() ModuleStats {
	stats := make(ModuleStats)
	buf := make([]byte, 1<<20)
	n := runtime.Stack(buf, true)
	stackData := buf[:n]

	sep := []byte("\n\n")
	goroutines := bytes.Split(stackData, sep)

	for _, goroutine := range goroutines {
		module := extractModule(goroutine)
		stats[module]++
	}

	return stats
}

func extractModule(goroutineStack []byte) string {
	lines := bytes.Split(goroutineStack, []byte("\n"))
	for i := 1; i < len(lines); i += 2 {
		line := strings.TrimSpace(string(lines[i]))
		module := parseModuleFromStackLine(line)
		if module != "" {
			return module
		}
	}
	return "other"
}

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
			parts := strings.Split(mod, "/")
			return parts[len(parts)-1]
		}
	}

	if strings.Contains(line, "runtime.") {
		return "runtime"
	}

	return ""
}
