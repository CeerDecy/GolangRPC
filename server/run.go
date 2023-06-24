package main

import (
	"fmt"
	"github/CeerDecy/RpcFrameWork/crpc"
	"github/CeerDecy/RpcFrameWork/server/models"
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
	group.UseMiddleWare(func(next crpc.HandleFunc) crpc.HandleFunc {
		return func(ctx *crpc.Context) {
			fmt.Println("pre handler")
			next(ctx)
		}
	})
	// **通配符
	group.Get("/hello/**", func(ctx *crpc.Context) {
		fmt.Fprintf(ctx.Writer, "%s hello get", "CeerDecy")
		fmt.Println("handler ... ")
	}, Log)

	// html
	group.Get("/html", func(ctx *crpc.Context) {
		ctx.HTML(http.StatusOK, "<h1>This is a html page</h1>")
	})

	// index模板
	group.Get("/html/index", func(ctx *crpc.Context) {
		ctx.HTMLTemplate("index.html", "", "static/html/index.html")
	})

	// login模板
	group.Get("/html/login", func(ctx *crpc.Context) {
		user := &models.User{Name: "猛喝威士忌"}
		//ctx.HTMLTemplate("login.html", user,
		//	"static/html/login.html", "static/html/header.html")
		ctx.HTMLTemplateGlob("login.html", user,
			"static/html/*.html")
	})
	engine.Run(":8000")
}
