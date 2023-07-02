package main

import (
	"errors"
	"fmt"
	"github/CeerDecy/RpcFrameWork/crpc"
	crpc_error "github/CeerDecy/RpcFrameWork/crpc/error"
	"github/CeerDecy/RpcFrameWork/crpc/pool"
	"github/CeerDecy/RpcFrameWork/crpc/token"
	"github/CeerDecy/RpcFrameWork/server/models"
	"log"
	"net/http"
	"sync"
	"time"
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
	engine := crpc.DefaultEngine()
	engine.UseMiddleWare(crpc.Logging, crpc.Recovery)
	engine.RegisterErrorHandler(func(err error) (int, any) {
		data := map[string]any{
			"code":  -999,
			"msg":   "err",
			"error": err.Error(),
		}
		return http.StatusInternalServerError, data
	})
	logger := engine.Logger
	// 加载HTML
	engine.LoadTemplate("static/html/*.html")
	group := engine.CreateGroup("user")
	group.UseMiddleWare(func(next crpc.HandleFunc) crpc.HandleFunc {
		return func(ctx *crpc.Context) {

			ctx.Logger.Info("MiddleWare", "log pre handler")
			next(ctx)
			ctx.Logger.Info("MiddleWare", ctx.Writer.Header().Get("Content-Type"))
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

	// 创建一个Basic账户认证的中间件
	ware := crpc.NewAccountMiddleWare(nil)
	ware.Users["猛喝威士忌"] = "123456"
	//fmt.Println(crpc.BasicAuth("猛喝威士忌", "123456"))
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

	group.Get("/recovery", func(ctx *crpc.Context) {
		panic("this is recovery request")
	})

	group.Get("/crError", func(ctx *crpc.Context) {
		crError := crpc_error.Default()
		crError.Result(func(c *crpc_error.CrError) {
			logger.Error("CrError", c.Error())
		})
		a(1, crError)
		ctx.JSON(http.StatusOK, nil)
	})

	group.Get("/httpErr", func(ctx *crpc.Context) {
		err := errors.New("http error")
		ctx.HandleWithError(err)
	})

	p, _ := pool.NewPoolConf()
	group.Get("/pool", func(ctx *crpc.Context) {
		currentTime := time.Now()
		var wg sync.WaitGroup
		wg.Add(6)
		_ = p.Submit(func() {
			defer wg.Done()
			fmt.Println("====6====")
			time.Sleep(2 * time.Second)
			panic("6 panic")
		})
		_ = p.Submit(func() {
			defer wg.Done()
			fmt.Println("====2====")
			time.Sleep(3 * time.Second)
		})
		_ = p.Submit(func() {
			defer wg.Done()
			fmt.Println("====3====")
			time.Sleep(2 * time.Second)
		})
		_ = p.Submit(func() {
			defer wg.Done()
			fmt.Println("====4====")
			time.Sleep(3 * time.Second)
		})
		_ = p.Submit(func() {
			defer wg.Done()
			fmt.Println("====5====")
			time.Sleep(3 * time.Second)
		})
		_ = p.Submit(func() {
			defer wg.Done()
			fmt.Println("====1====")
			time.Sleep(2 * time.Second)
		})
		//p.Release()
		wg.Wait()
		logger.Info("Pool", fmt.Sprintf("start time:%v , end time:%v", currentTime.Format("15:04:05"), time.Now().Format("15:04:05")))
		ctx.JSON(http.StatusOK, "success")
	})

	group.Get("/login", func(ctx *crpc.Context) {
		jwt := &token.JwtHandler{}
		jwt.Key = []byte("123456")
		jwt.SendCookie = true
		jwt.TimeOut = 10 * time.Minute
		jwt.RefreshTimeOut = 20 * time.Minute
		jwt.Authenticator = func(ctx *crpc.Context) (map[string]any, error) {
			data := make(map[string]any)
			data["userId"] = 1
			return data, nil
		}
		handler, err := jwt.LoginHandler(ctx)
		if err != nil {
			ctx.Logger.Error("/login", err)
			ctx.JSON(http.StatusOK, err)
			return
		}
		ctx.JSON(http.StatusOK, handler)
	})

	jwt := &token.JwtHandler{}
	group.Get("/refresh", func(ctx *crpc.Context) {
		jwt.Key = []byte("123456")
		jwt.SendCookie = true
		jwt.TimeOut = 10 * time.Minute
		jwt.RefreshTimeOut = 20 * time.Minute
		jwt.RefreshKey = "server_refresh_token"
		ctx.Set(jwt.RefreshKey, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODgyMjIzNDYsImlhdCI6MTY4ODIyMTE0NiwidXNlcklkIjoxfQ.WxpzEfe0QIrR_84CLnazqybNFZjItV6aNbUmRb18PYY")
		jwt.Authenticator = func(ctx *crpc.Context) (map[string]any, error) {
			data := make(map[string]any)
			data["userId"] = 1
			return data, nil
		}
		handler, err := jwt.RefreshHandler(ctx)
		ctx.Logger.Info("/refresh", "Okay")
		fmt.Println(err)
		if err != nil {
			fmt.Println(err)
			logger.Error("/refresh", err)
			ctx.Logger.Error("/refresh", err)
			ctx.JSON(http.StatusOK, map[string]any{
				"error": err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, handler)
	}, jwt.AuthInterceptor)
	//engine.Run(":8000")
	engine.RunTLS(":8080", "key/server.pem", "key/server.key")
}

func a(n int, crErr *crpc_error.CrError) {
	if n == 1 {
		err := errors.New("a error")
		crErr.Put(err)
	}
}
