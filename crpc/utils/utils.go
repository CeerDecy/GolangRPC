package utils

import (
	"fmt"
	"reflect"
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

func JoinStrings(str ...any) string {
	var builder strings.Builder
	for i := range str {
		builder.WriteString(checkStr(str[i]))
	}
	return builder.String()
}

func checkStr(a any) string {
	if reflect.ValueOf(a).Kind() == reflect.String {
		return a.(string)
	} else {
		return fmt.Sprintf("%v", a)
	}
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
