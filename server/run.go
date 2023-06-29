package main

import (
	"fmt"
	"github/CeerDecy/RpcFrameWork/crpc"
	"github/CeerDecy/RpcFrameWork/crpc/crpcLogger"
	"github/CeerDecy/RpcFrameWork/server/models"
	"log"
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
	logger := crpcLogger.TextLogger()
	logger.WriterInFile("./log/")
	logger.LogFileSize = 1 << 10
	// 初始化引擎
	engine := crpc.MakeEngine()
	// 加载HTML
	engine.LoadTemplate("static/html/*.html")
	group := engine.CreateGroup("user")
	group.UseMiddleWare(crpc.Logging)
	group.UseMiddleWare(func(next crpc.HandleFunc) crpc.HandleFunc {
		return func(ctx *crpc.Context) {
			logger.Debug("MiddleWare", "log pre handler")
			next(ctx)
			logger.Info("MiddleWare", "log post handler")
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

	// 获取参数Map
	group.Get("/queryMap", func(ctx *crpc.Context) {
		queryMap, _ := ctx.GetQueryMap("user")
		ctx.JSON(http.StatusOK, queryMap)
	})

	// Post表单
	group.Get("/postForm", func(ctx *crpc.Context) {
		res, _ := ctx.GetPostFormMap("userInfo")
		ctx.JSON(http.StatusOK, res)
	})

	// 文件上传
	group.Get("/fileUpload", func(ctx *crpc.Context) {
		file := ctx.FormFile("file")
		err := ctx.SaveUploadFile(file, "./upload/"+file.Filename)
		if err != nil {
			log.Println(err)
		}
	})

	// Json RequestBody参数
	group.Get("/jsonParam", func(ctx *crpc.Context) {
		user := &models.User{}
		err := ctx.BindJson(&user)
		if err != nil {
			log.Println(err)
			ctx.JSON(http.StatusOK, map[string]any{
				"error": err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, user)
	})

	// XML RequestBody参数
	group.Any("/xmlParam", func(ctx *crpc.Context) {
		user := &models.User{}
		err := ctx.BindXML(&user)
		if err != nil {
			log.Println(err)
			dict := map[string]any{
				"error": err.Error(),
			}
			ctx.JSON(http.StatusInternalServerError, dict)
			logger.Error(ctx.Request.RequestURI, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, user)
	})
	engine.Run(":8000")
}
