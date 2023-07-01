package crpc

import (
	"encoding/base64"
	"net/http"
)

type AccountMiddleWare struct {
	UnAuthHandler func(ctx *Context)
	Users         map[string]string
}

func NewAccountMiddleWare(handler func(ctx *Context)) *AccountMiddleWare {
	return &AccountMiddleWare{
		UnAuthHandler: handler,
		Users:         make(map[string]string),
	}
}

func (a *AccountMiddleWare) BasicAuth(next HandleFunc) HandleFunc {
	return func(ctx *Context) {
		// 从Header中获取Base64字符串
		username, password, ok := ctx.Request.BasicAuth()
		if !ok {
			a.unAuthHandler(ctx)
			return
		}
		if pwd, ok := a.Users[username]; !ok {
			a.unAuthHandler(ctx)
			return
		} else if pwd != password {
			a.unAuthHandler(ctx)
			return
		}
		ctx.Set("user", username)
		next(ctx)
	}
}

// 判断UnAuthHandler是否为空，若不为空则执行处理函数，为空就执行默认处理
func (a *AccountMiddleWare) unAuthHandler(ctx *Context) {
	if a.UnAuthHandler != nil {
		a.UnAuthHandler(ctx)
		return
	}
	ctx.JSON(http.StatusUnauthorized, map[string]any{
		"code": http.StatusUnauthorized,
		"msg":  "un authorized",
	})
}

func BasicAuth(username string, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
