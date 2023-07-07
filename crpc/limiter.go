package crpc

import (
	"golang.org/x/net/context"
	"golang.org/x/time/rate"
	"net/http"
	"time"
)

func Limiter(limit, cap int) MiddleWareFunc {
	limiter := rate.NewLimiter(rate.Limit(limit), cap)
	return func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {

			timeout, cancelFunc := context.WithTimeout(context.Background(), time.Duration(1)*time.Second)
			defer cancelFunc()
			err := limiter.WaitN(timeout, 1)
			if err != nil {
				ctx.String(http.StatusForbidden, "被限流了")
				return
			}
			next(ctx)

		}
	}
}
