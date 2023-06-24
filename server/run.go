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
	group.PreMiddleWare(func(next crpc.HandleFunc) crpc.HandleFunc {
		return func(ctx *crpc.Context) {
			fmt.Println("pre handler")
			next(ctx)
		}
	})
	group.Get("/hello/**", func(ctx *crpc.Context) {
		fmt.Fprintf(ctx.Writer, "%s hello get", "CeerDecy")
	})
	engine.Run(":8000")
}
