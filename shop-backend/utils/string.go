package utils

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// RuneLen 返回字符串的字符数（按 Unicode 码点计算）
// 与 len() 不同，len() 返回字节数，对中文/emoji等多字节字符会偏大
// 对应 MySQL varchar(N) 的字符数限制
func RuneLen(s string) int {
	return utf8.RuneCountInString(s)
}

// TruncateByRune 按字符数截断字符串，超过 maxLen 个字符则截断
// 保证截断后的字符数 <= maxLen
func TruncateByRune(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxLen])
}

// IsRuneLenInRange 检查字符串字符数是否在 [min, max] 范围内
func IsRuneLenInRange(s string, min, max int) bool {
	n := utf8.RuneCountInString(s)
	return n >= min && n <= max
}

// IsRuneLenMax 检查字符串字符数是否不超过 maxLen
func IsRuneLenMax(s string, maxLen int) bool {
	return utf8.RuneCountInString(s) <= maxLen
}

// IsEmpty 检查字符串是否为空（去除空白后为空）
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotEmpty 检查字符串是否非空（去除空白后非空）
func IsNotEmpty(s string) bool {
	return strings.TrimSpace(s) != ""
}

// ContainsChinese 检查字符串是否包含中文字符
func ContainsChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

// ContainsMultiByte 检查字符串是否包含多字节字符（中文、emoji、特殊符号等）
// 若包含，则 len(s) != utf8.RuneCountInString(s)
func ContainsMultiByte(s string) bool {
	return len(s) != utf8.RuneCountInString(s)
}
