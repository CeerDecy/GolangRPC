package main

import (
	"fmt"
	"github/CeerDecy/RpcFrameWork/crpc"
	"net/http"
)

func Log(next crpc.HandleFunc) crpc.HandleFunc {
	return func(ctx *crpc.Context) {
		fmt.Println("router pre handler -> ", ctx.Request.RequestURI)
		next(ctx)
		fmt.Println("router post handler")
	}
}

func main() {
	engine := crpc.MakeEngine()
	group := engine.CreateGroup("user")
	//group.Post("/hello", func(ctx *crpc.Context) {
	//	fmt.Fprintf(ctx.Writer, "%s hello post", "CeerDecy")
	//})
	group.UseMiddleWare(func(next crpc.HandleFunc) crpc.HandleFunc {
		return func(ctx *crpc.Context) {
			fmt.Println("pre handler")
			next(ctx)
		}
	})
	//group.PostMiddleWare(func(pre crpc.HandleFunc) crpc.HandleFunc {
	//	return func(ctx *crpc.Context) {
	//		fmt.Println("this is post handler")
	//	}
	//})
	group.Get("/hello/**", func(ctx *crpc.Context) {
		fmt.Fprintf(ctx.Writer, "%s hello get", "CeerDecy")
		fmt.Println("handler ... ")
	}, Log)

	group.Get("/html", func(ctx *crpc.Context) {
		ctx.HTML(http.StatusOK, "<h1>This is a html page</h1>")
	})
	engine.Run(":8000")
}
