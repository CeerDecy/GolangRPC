package crpc

import (
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github/CeerDecy/RpcFrameWork/crpc/config"
	"github/CeerDecy/RpcFrameWork/crpc/crpcLogger"
	"github/CeerDecy/RpcFrameWork/crpc/gateway"
	"github/CeerDecy/RpcFrameWork/crpc/register"
	"github/CeerDecy/RpcFrameWork/crpc/render"
	"github/CeerDecy/RpcFrameWork/crpc/utils"
	"html/template"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

const MethodAny = "MethodAny"

// HandleFunc 请求处理函数
type HandleFunc func(ctx *Context)

// MiddleWareFunc 中间件执行函数
type MiddleWareFunc func(next HandleFunc) HandleFunc

// 路由组
type routerGroup struct {
	groupName          string                                 // 组名
	HandleFuncMap      map[string]map[string]HandleFunc       // 组中对应的路由方法
	middleWaresFuncMap map[string]map[string][]MiddleWareFunc // 组中对应的路由方法
	treeNode           treeNode                               // 路径前缀树
	middleWares        []MiddleWareFunc                       // 中间件
	//postMiddleWares []MiddleWareFunc                 // 后置中间件
}

// UseMiddleWare 添加前置中间件
func (group *routerGroup) UseMiddleWare(wareFunc ...MiddleWareFunc) {
	group.middleWares = append(group.middleWares, wareFunc...)
}

//// PostMiddleWare 添加后置中间件
//func (group *routerGroup) PostMiddleWare(wareFunc ...MiddleWareFunc) {
//	group.postMiddleWares = append(group.postMiddleWares, wareFunc...)
//}

func (group *routerGroup) methodHandle(path string, method string, handleFunc HandleFunc, ctx *Context) {
	// 前置中间件 -> 组级别通用中间件
	if group.middleWares != nil {
		for _, middleWareFunc := range group.middleWares {
			handleFunc = middleWareFunc(handleFunc)
		}
	}
	// 路由级别中间件
	if group.middleWaresFuncMap[path][method] != nil {
		for _, middleWareFunc := range group.middleWaresFuncMap[path][method] {
			handleFunc = middleWareFunc(handleFunc)
		}
	}
	handleFunc(ctx)
	// 后置中间件
	//if group.postMiddleWares != nil {
	//	for _, middleWareFunc := range group.postMiddleWares {
	//		handleFunc = middleWareFunc(handleFunc)
	//	}
	//}
	//handleFunc(ctx)
}

func (group *routerGroup) handle(route, method string, handleFunc HandleFunc, middleware ...MiddleWareFunc) {
	if _, ok := group.HandleFuncMap[route]; !ok {
		group.HandleFuncMap[route] = make(map[string]HandleFunc)
		group.middleWaresFuncMap[route] = make(map[string][]MiddleWareFunc)
	}
	if _, ok := group.HandleFuncMap[route][method]; ok {
		panic("this crpc has exist")
	}
	group.HandleFuncMap[route][method] = handleFunc
	group.middleWaresFuncMap[route][method] = append(group.middleWaresFuncMap[route][method], middleware...)
	group.treeNode.Put(route)
}

// Any 为当前组别添加路由方法
func (group *routerGroup) Any(route string, handleFunc HandleFunc, middleware ...MiddleWareFunc) {
	group.handle(route, MethodAny, handleFunc, middleware...)
}

// Get 配置Get请求的路由
func (group *routerGroup) Get(route string, handleFunc HandleFunc, middleware ...MiddleWareFunc) {
	group.handle(route, http.MethodGet, handleFunc, middleware...)
}

// Post 配置Post请求的路由
func (group *routerGroup) Post(route string, handleFunc HandleFunc, middleware ...MiddleWareFunc) {
	group.handle(route, http.MethodPost, handleFunc, middleware...)
}

// Delete 配置Delete请求的路由
func (group *routerGroup) Delete(route string, handleFunc HandleFunc, middleware ...MiddleWareFunc) {
	group.handle(route, http.MethodDelete, handleFunc, middleware...)
}

// Put 配置Put请求的路由
func (group *routerGroup) Put(route string, handleFunc HandleFunc, middleware ...MiddleWareFunc) {
	group.handle(route, http.MethodPut, handleFunc, middleware...)
}

// Patch 配置Patch请求的路由
func (group *routerGroup) Patch(route string, handleFunc HandleFunc, middleware ...MiddleWareFunc) {
	group.handle(route, http.MethodPatch, handleFunc, middleware...)
}

// Options 配置Patch请求的路由
func (group *routerGroup) Options(route string, handleFunc HandleFunc, middleware ...MiddleWareFunc) {
	group.handle(route, http.MethodOptions, handleFunc, middleware...)
}

// Head 配置Patch请求的路由
func (group *routerGroup) Head(route string, handleFunc HandleFunc, middleware ...MiddleWareFunc) {
	group.handle(route, http.MethodHead, handleFunc, middleware...)
}

// 用于存储路由表
type router struct {
	RouterGroups []*routerGroup
	engine       *Engine
}

// CreateGroup 添加组别
func (r *router) CreateGroup(groupName string) *routerGroup {
	group := &routerGroup{
		groupName:          groupName,
		HandleFuncMap:      make(map[string]map[string]HandleFunc),
		middleWaresFuncMap: make(map[string]map[string][]MiddleWareFunc),
		//HandleMethodMap: make(map[string][]string),
		treeNode: treeNode{name: groupName},
	}
	group.middleWares = r.engine.middles
	r.RouterGroups = append(r.RouterGroups, group)
	return group
}

// ErrorHandler HTTP错误处理函数
type ErrorHandler func(err error) (int, any)

type Engine struct {
	router
	funcMap          template.FuncMap
	HTMLRender       *render.HTMLRender
	Pool             sync.Pool
	Logger           *crpcLogger.Logger
	middles          []MiddleWareFunc
	errorHandler     ErrorHandler
	OpenGateway      bool
	gatewayConfigs   []*gateway.GWConfig
	gatewayTreeNode  *gateway.TreeNode
	gatewayConfigMap map[string]*gateway.GWConfig
	RegClient        naming_client.INamingClient
}

// MakeEngine 初始化引擎
func MakeEngine() *Engine {
	e := &Engine{
		router: router{},
		gatewayTreeNode: &gateway.TreeNode{
			Name:  "/",
			Child: make([]*gateway.TreeNode, 0),
		},
		gatewayConfigMap: make(map[string]*gateway.GWConfig),
	}
	e.Pool.New = func() any {
		return e.allocateContext()
	}
	return e
}

func DefaultEngine() *Engine {
	engine := MakeEngine()
	engine.Logger = crpcLogger.TextLogger()
	if path, ok := config.Conf.Log["path"]; ok {
		if size, ok := config.Conf.Log["size"]; ok {
			engine.Logger.WriterInFile(path.(string))
			engine.Logger.LogFileSize = size.(int64) << 20
		} else {
			log.Fatalln("log size config is null")
		}
	}
	engine.router.engine = engine
	return engine
}

// SetGatewayConfig 设置gateway配置
func (e *Engine) SetGatewayConfig(conf []*gateway.GWConfig) {
	e.gatewayConfigs = append(e.gatewayConfigs, conf...)
	for index, v := range e.gatewayConfigs {
		e.gatewayTreeNode.Put(v.Path, v.Name)
		e.gatewayConfigMap[v.Name] = e.gatewayConfigs[index]
	}
}

// 给Context分配内存
func (e *Engine) allocateContext() any {
	return &Context{
		engine: e,
	}
}

// SetFuncMap 设置FuncMap
func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

// LoadTemplate 根据路径通配符加载模板
func (e *Engine) LoadTemplate(pattern string) {
	t := template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
	e.SetTemplate(t)
}

// SetTemplate 用户自定义设置模板
func (e *Engine) SetTemplate(t *template.Template) {
	e.HTMLRender = &render.HTMLRender{Template: t}
}

// 实现Handler接口
func (e *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := e.Pool.Get().(*Context)
	ctx.Writer = writer
	ctx.Request = request
	ctx.Logger = e.Logger
	e.HttpRequestHandle(ctx, writer, request)
	e.Pool.Put(ctx)
}

// Run 启动引擎
func (e *Engine) Run(address string) {
	client, err := register.CreateNacosClient()
	if err != nil {
		e.Logger.Debug("run register", err)
	}
	e.RegClient = client
	http.Handle("/", e)
	err = http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (e *Engine) HttpRequestHandle(ctx *Context, writer http.ResponseWriter, request *http.Request) {
	if e.OpenGateway {
		// 网关处理逻辑
		path := request.URL.Path
		node := e.gatewayTreeNode.Get(path)
		if node == nil {
			writer.WriteHeader(http.StatusNotFound)
			ctx.Logger.Error("Gateway", "404 not found")
			return
		}
		fmt.Println(node.Name)
		gwConfig := e.gatewayConfigMap[node.GwName]
		instance, port, err := register.GetInstance(e.RegClient, gwConfig.ServiceName)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			ctx.Logger.Error("Gateway", err.Error())
			return
		}
		fmt.Println(fmt.Sprintf("http://%s:%d%s", instance, port, path))
		target, err := url.Parse(fmt.Sprintf("http://%s:%d%s", instance, port, path))
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			ctx.Logger.Error("Gateway", err.Error())
			return
		}
		// 重定向请求
		director := func(request *http.Request) {
			request.Host = target.Host
			request.URL.Host = target.Host
			request.URL.Path = target.Path
			request.URL.Scheme = target.Scheme
			if _, ok := request.Header["User-Agent"]; !ok {
				request.Header.Set("User-Agent", "")
			}
			if gwConfig.Header != nil {
				gwConfig.Header(request)
			}
			fmt.Println("请求")
		}
		response := func(response *http.Response) error {
			fmt.Println("响应修改")
			return nil
		}
		handler := func(writer http.ResponseWriter, request *http.Request, err error) {
			e.Logger.Error("Gateway Error", err.Error())
			fmt.Println("错误处理")
		}
		proxy := httputil.ReverseProxy{Director: director, ModifyResponse: response, ErrorHandler: handler}
		proxy.ServeHTTP(writer, request)
		return
	}
	// 获取当前请求的方法
	method := request.Method
	// 遍历Group
	for _, group := range e.RouterGroups {
		routeName := utils.SubStringLast(request.URL.Path, "/"+group.groupName)
		node := group.treeNode.Get(routeName)
		if node != nil && node.isEnd {
			// 若请求方式为Any，则直接运行Any中的方法
			if handleFunc, ok := group.HandleFuncMap[node.routePath][MethodAny]; ok {
				group.methodHandle(node.routePath, MethodAny, handleFunc, ctx)
				return
			}
			// 若请求方式为method，那么执行method中的方法
			if handleFunc, ok := group.HandleFuncMap[node.routePath][method]; ok {
				group.methodHandle(node.routePath, method, handleFunc, ctx)
				return
			}
			// 执行到这说明当前路由请求的方法不被服务器所支持
			writer.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = fmt.Fprintf(writer, "%s is not allowed", request.RequestURI)
			return
		}
	}
	// 遍历完成
	writer.WriteHeader(http.StatusNotFound)
	_, _ = writer.Write([]byte("404 " + request.RequestURI + " resource not found"))
}

// UseMiddleWare 配置中间件
func (e *Engine) UseMiddleWare(middles ...MiddleWareFunc) {
	e.middles = middles
}

func (e *Engine) RegisterErrorHandler(handler ErrorHandler) {
	e.errorHandler = handler
}

// RunTLS 开启HTTPS
func (e *Engine) RunTLS(addr, certFile, keyFile string) {
	err := http.ListenAndServeTLS(addr, certFile, keyFile, e.Handler())
	if err != nil {
		log.Fatalln(err)
	}
}

func (e *Engine) Handler() http.Handler {
	return e
}
