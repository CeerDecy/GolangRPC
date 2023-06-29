package crpc

import (
	"errors"
	"fmt"
	crpc_error "github/CeerDecy/RpcFrameWork/crpc/error"
	"net/http"
	"runtime"
	"strings"
)

func detailMsg(err any) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:])
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%v", err))
	for _, pc := range pcs[0:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		builder.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return builder.String()
}

// Recovery 中间件
func Recovery(next HandleFunc) HandleFunc {
	return func(ctx *Context) {
		defer func() {
			if err := recover(); err != nil {
				if e := err.(error); e != nil {
					var crErr *crpc_error.CrError
					if errors.As(e, &crErr) {
						crErr.ExecResult()
						return
					}
				}
				ctx.Logger.Error("Recovery", detailMsg(err))
				ctx.Fail(http.StatusInternalServerError, err)
			}
		}()
		next(ctx)
	}
}
