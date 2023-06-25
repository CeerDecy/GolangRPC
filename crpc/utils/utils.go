package utils

import (
	"strings"
	"unicode"
)

func SubStringLast(str string, substr string) string {
	index := strings.Index(str, substr)
	if index < 0 {
		return ""
	}
	return str[index+len(substr):]
}

// IsASCII 判断字符是否为ASCII编码
func IsASCII(str string) bool {
	for i := range str {
		if str[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}
