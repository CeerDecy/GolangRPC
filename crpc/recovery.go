package crpc

import "net/http"

func Recovery(next HandleFunc) HandleFunc {
	return func(ctx *Context) {
		defer func() {
			if err := recover(); err != nil {
				ctx.Logger.Error("Recovery", err)
				ctx.Fail(http.StatusInternalServerError, err)
			}
		}()
		next(ctx)
	}
}
