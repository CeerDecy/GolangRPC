package crpc

import (
	"fmt"
	"github/CeerDecy/RpcFrameWork/crpc/utils"
	"log"
	"net/http"
)

const MethodAny = "MethodAny"

// HandleFunc 请求处理函数
type HandleFunc func(ctx *Context)

// MiddleWareFunc 中间件执行函数
type MiddleWareFunc func(handleFunc HandleFunc) HandleFunc

// 路由组
type routerGroup struct {
	groupName       string                           // 组名
	HandleFuncMap   map[string]map[string]HandleFunc // 组中对应的路由方法
	treeNode        treeNode                         // 路径前缀树
	preMiddleWares  []MiddleWareFunc                 // 前置中间件
	postMiddleWares []MiddleWareFunc                 // 后置中间件
}

// PreMiddleWare 添加前置中间件
func (group *routerGroup) PreMiddleWare(wareFunc ...MiddleWareFunc) {
	group.preMiddleWares = append(group.preMiddleWares, wareFunc...)
}

// PostMiddleWare 添加后置中间件
func (group *routerGroup) PostMiddleWare(wareFunc ...MiddleWareFunc) {
	group.postMiddleWares = append(group.postMiddleWares, wareFunc...)
}

func (group *routerGroup) methodHandle(handleFunc HandleFunc, ctx *Context) {
	// 前置中间件
	if group.preMiddleWares != nil {
		for _, middleWareFunc := range group.preMiddleWares {
			handleFunc = middleWareFunc(handleFunc)
		}
	}
	handleFunc(ctx)
	// 后置中间件
	if group.postMiddleWares != nil {
		for _, middleWareFunc := range group.postMiddleWares {
			handleFunc = middleWareFunc(handleFunc)
		}
	}
}

func (group *routerGroup) handle(route, method string, handleFunc HandleFunc) {
	if _, ok := group.HandleFuncMap[route]; !ok {
		group.HandleFuncMap[route] = make(map[string]HandleFunc)
	}
	if _, ok := group.HandleFuncMap[route][method]; ok {
		panic("this crpc has exist")
	}
	group.HandleFuncMap[route][method] = handleFunc
	group.treeNode.Put(route)
}

// Any 为当前组别添加路由方法
func (group *routerGroup) Any(route string, handleFunc HandleFunc) {
	group.handle(route, MethodAny, handleFunc)
}

// Get 配置Get请求的路由
func (group *routerGroup) Get(route string, handleFunc HandleFunc) {
	group.handle(route, http.MethodGet, handleFunc)
}

// Post 配置Post请求的路由
func (group *routerGroup) Post(route string, handleFunc HandleFunc) {
	group.handle(route, http.MethodPost, handleFunc)
}

// Delete 配置Delete请求的路由
func (group *routerGroup) Delete(route string, handleFunc HandleFunc) {
	group.handle(route, http.MethodDelete, handleFunc)
}

// Put 配置Put请求的路由
func (group *routerGroup) Put(route string, handleFunc HandleFunc) {
	group.handle(route, http.MethodPut, handleFunc)
}

// Patch 配置Patch请求的路由
func (group *routerGroup) Patch(route string, handleFunc HandleFunc) {
	group.handle(route, http.MethodPatch, handleFunc)
}

// Options 配置Patch请求的路由
func (group *routerGroup) Options(route string, handleFunc HandleFunc) {
	group.handle(route, http.MethodOptions, handleFunc)
}

// Head 配置Patch请求的路由
func (group *routerGroup) Head(route string, handleFunc HandleFunc) {
	group.handle(route, http.MethodHead, handleFunc)
}

// 用于存储路由表
type router struct {
	RouterGroups []*routerGroup
}

// CreateGroup 添加组别
func (r *router) CreateGroup(groupName string) *routerGroup {
	group := &routerGroup{
		groupName:     groupName,
		HandleFuncMap: make(map[string]map[string]HandleFunc),
		//HandleMethodMap: make(map[string][]string),
		treeNode: treeNode{name: groupName},
	}
	r.RouterGroups = append(r.RouterGroups, group)
	return group
}

type Engine struct {
	router
}

// MakeEngine 初始化引擎
func MakeEngine() *Engine {
	return &Engine{
		router{},
	}
}

func (e *Engine) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	e.HttpRequestHandle(writer, request)

}

// Run 启动引擎
func (e *Engine) Run(address string) {
	http.Handle("/", e)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (e *Engine) HttpRequestHandle(writer http.ResponseWriter, request *http.Request) {
	// 获取当前请求的方法
	method := request.Method
	// 遍历Group
	for _, group := range e.RouterGroups {
		routeName := utils.SubStringLast(request.RequestURI, "/"+group.groupName)
		node := group.treeNode.Get(routeName)
		if node != nil && node.isEnd {
			// 若请求方式为Any，则直接运行Any中的方法
			if handleFunc, ok := group.HandleFuncMap[node.routePath][MethodAny]; ok {
				group.methodHandle(handleFunc, &Context{writer, request})
				return
			}
			// 若请求方式为method，那么执行method中的方法
			if handleFunc, ok := group.HandleFuncMap[node.routePath][method]; ok {
				group.methodHandle(handleFunc, &Context{writer, request})
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
