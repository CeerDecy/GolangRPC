package main

import (
	"fmt"
	"github/CeerDecy/RpcFrameWork/crpc"
)

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
			fmt.Println("post handler")
		}
	})
	group.UseMiddleWare(func(next crpc.HandleFunc) crpc.HandleFunc {
		return func(ctx *crpc.Context) {
			fmt.Println("two pre")
			next(ctx)
			fmt.Println("two post")
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
	})
	engine.Run(":8000")
}
