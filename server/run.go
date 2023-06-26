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
		fmt.Println("router post handler -> ", ctx.Request.RequestURI)
		fmt.Println(ctx.Writer.Header().Get("Content-Type"))
	}
}

func main() {
	// 初始化引擎
	engine := crpc.MakeEngine()
	// 加载HTML
	engine.LoadTemplate("static/html/*.html")
	group := engine.CreateGroup("user")
	group.UseMiddleWare(func(next crpc.HandleFunc) crpc.HandleFunc {
		return func(ctx *crpc.Context) {
			fmt.Println("pre handler")
			next(ctx)
			fmt.Println("post handler")
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
	// Register模板
	group.Get("/html/register", func(ctx *crpc.Context) {
		user := &models.User{Name: "猛喝威士忌"}
		ctx.Template("register.html", user)
	})

	// 返回JSON数据
	group.Get("/json", func(ctx *crpc.Context) {
		user := &models.User{Name: "猛喝威士忌"}
		ctx.JSON(http.StatusOK, user)
	})

	// 返回xml数据
	group.Get("/xml", func(ctx *crpc.Context) {
		user := &models.User{Name: "猛喝威士忌"}
		ctx.XML(http.StatusOK, user)
	})

	// 返回jpeg文件
	group.Get("/image", func(ctx *crpc.Context) {
		ctx.File("static/img/image.jpeg")
	})

	// 返回jpeg文件
	group.Get("/imageByName", func(ctx *crpc.Context) {
		ctx.FileAttachment("static/img/image.jpeg", "car.jpeg")
	})

	// 通过文件系统获取文件
	group.Get("/filesystem", func(ctx *crpc.Context) {
		ctx.FileFromFS("image.jpeg", http.Dir("static/img/"))
	})

	// 重定向
	group.Get("/redirect", func(ctx *crpc.Context) {
		// 重定向的状态值为302
		ctx.Redirect(http.StatusFound, "/user/html/login")
	})

	// 格式化字符串
	group.Get("/string", func(ctx *crpc.Context) {
		ctx.String(http.StatusOK, "%s %s", "CeerDecy", "猛喝威士忌")
	})

	// 获取参数
	group.Get("/add", func(ctx *crpc.Context) {
		ctx.String(http.StatusOK, ctx.GetDefaultQuery("a", "zzz"))
	})

	engine.Run(":8000")
}
