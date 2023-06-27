package crpc_error

import (
	"fmt"
	"strings"
)

type SliceValidateError []error

func (e SliceValidateError) Error() string {
	n := len(e)
	switch n {
	case 0:
		return ""
	default:
		var builder strings.Builder
		if e[0] != nil {
			fmt.Fprintf(&builder, "[%d]:%s", 0, e[0].Error())
		}
		if n > 1 {
			for i := 1; i < n; i++ {
				if e[i] != nil {
					builder.WriteString("\n")
					fmt.Fprintf(&builder, "[%d]:%s", i, e[i].Error())
				}
			}
		}
		return builder.String()
	}
}
