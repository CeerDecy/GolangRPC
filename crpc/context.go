package crpc

import (
	"errors"
	"fmt"
	"github/CeerDecy/RpcFrameWork/crpc/render"
	"github/CeerDecy/RpcFrameWork/crpc/utils"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const defaultMaxMemory = 32 << 20 // 32 MB

type Context struct {
	Writer     http.ResponseWriter
	Request    *http.Request
	engine     *Engine
	queryCache url.Values
	formCache  url.Values
}

// FormFile 获取表单中的文件
func (c *Context) FormFile(name string) *multipart.FileHeader {
	file, header, err := c.Request.FormFile(name)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	return header
}

// FormFiles 获取表单中的多个文件
func (c *Context) FormFiles(name string) []*multipart.FileHeader {
	multipartForm, err := c.MultipartForm()
	if err != nil {
		log.Println(err)
	}
	return multipartForm.File[name]
}

// SaveUploadFile 保存上传的文件
func (c *Context) SaveUploadFile(file *multipart.FileHeader, dst string) error {
	open, err := file.Open()
	if err != nil {
		return err
	}
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, err = io.Copy(d, open)
	return err
}

func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.Request.ParseMultipartForm(defaultMaxMemory)
	return c.Request.MultipartForm, err
}

// 初始化Post表单参数
func (c *Context) initPostFormCache() {
	if c.Request != nil {
		if err := c.Request.ParseMultipartForm(defaultMaxMemory); err != nil {
			if !errors.Is(err, http.ErrNotMultipart) {
				log.Println(err)
			}
		}

		c.formCache = c.Request.PostForm
	} else {
		c.formCache = url.Values{}
	}
}

// GetPostFormArray 获取参数
func (c *Context) GetPostFormArray(key string) ([]string, bool) {
	c.initPostFormCache()
	val, ok := c.formCache[key]
	return val, ok
}

// GetPostFormMap 获取参数
func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	c.initPostFormCache()
	return c.get(c.formCache, key)
}

func (c *Context) PostFormArray(key string) []string {
	values, _ := c.GetPostFormArray(key)
	return values
}

// GetPostForm 获取Map中的value
func (c *Context) GetPostForm(key string) (string, bool) {
	values, ok := c.GetPostFormArray(key)
	if ok {
		return values[0], ok
	}
	return "", false
}

// 初始化参数缓存
func (c *Context) initQueryCache() {
	if c.Request != nil {
		c.queryCache = c.Request.URL.Query()
	} else {
		c.queryCache = url.Values{}
	}
}

// GetQuery 获取参数
func (c *Context) GetQuery(key string) string {
	c.initQueryCache()
	return c.queryCache.Get(key)
}

// GetQueryArray 获取参数
func (c *Context) GetQueryArray(key string) ([]string, bool) {
	c.initQueryCache()
	val, ok := c.queryCache[key]
	return val, ok
}

// GetDefaultQuery 获取参数
func (c *Context) GetDefaultQuery(key, def string) string {
	val, ok := c.GetQueryArray(key)
	if !ok {
		return def
	}
	return val[0]
}

// GetQueryMap 获取请求中Map参数
// http://172.0.0.1:8000/user/queryMap?user[name]=ABC&user[age]=18
func (c *Context) GetQueryMap(key string) (map[string]string, bool) {
	c.initQueryCache()
	return c.get(c.queryCache, key)
}

// QueryMap 获取请求中的参数Map
func (c *Context) QueryMap(key string) map[string]string {
	dict, _ := c.GetQueryMap(key)
	return dict
}

// 获取指定Map中的所有键值对
func (c *Context) get(cache map[string][]string, key string) (map[string]string, bool) {
	dict := make(map[string]string)
	exist := false
	for k, v := range cache {
		left := strings.IndexByte(k, '[')
		right := strings.IndexByte(k, ']')
		if left >= 1 && right >= 1 && k[:left] == key {
			exist = true
			dict[k[left+1:right]] = v[0]
		}
	}
	return dict, exist
}

// HTML 返回HTML文本
func (c *Context) HTML(status int, html string) {
	c.Render(status, &render.HTML{Data: html, IsTemp: false})
}

// HTMLTemplate 返回HTML模板
func (c *Context) HTMLTemplate(name string, data any, filename ...string) {
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	temp := template.New(name)
	temp, err := temp.ParseFiles(filename...)
	if err != nil {
		log.Println(err)
		return
	}
	err = temp.Execute(c.Writer, data)
	if err != nil {
		log.Println(err)
	}
}

// HTMLTemplateGlob 返回HTML模板
// 以通配符的形式加载html文件
func (c *Context) HTMLTemplateGlob(name string, data any, filepath string) {
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	temp := template.New(name)
	temp, err := temp.ParseGlob(filepath)
	if err != nil {
		log.Println(err)
		return
	}
	err = temp.Execute(c.Writer, data)
	if err != nil {
		log.Println(err)
	}
}

// Template 加载Template
func (c *Context) Template(name string, data any) {
	c.Render(http.StatusOK, &render.HTML{
		Name:   name,
		Data:   data,
		Temp:   c.engine.HTMLRender.Template,
		IsTemp: true,
	})
}

// JSON 返回JSON数据
func (c *Context) JSON(state int, data any) {
	c.Render(state, &render.Json{Data: data})
}

// XML 返回XML数据
func (c *Context) XML(state int, data any) {
	c.Render(state, &render.XML{Data: data})
	//log.Println(c.Writer.Header().Get("Content-Type"))
}

// File 返回文件数据
func (c *Context) File(filename string) {
	http.ServeFile(c.Writer, c.Request, filename)
}

// FileAttachment 返回文件数据
func (c *Context) FileAttachment(filepath, filename string) {
	if utils.IsASCII(filename) {
		c.Writer.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	} else {
		c.Writer.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename*="UTF-8''%s`, url.QueryEscape(filename)))
	}
	http.ServeFile(c.Writer, c.Request, filepath)
}

// FileFromFS 从文件系统中获取文件
// filepath 是相对于文件系统的路径
func (c *Context) FileFromFS(filepath string, fs http.FileSystem) {
	defer func(old string) {
		c.Request.URL.Path = old
	}(c.Request.URL.Path)
	c.Request.URL.Path = filepath
	http.FileServer(fs).ServeHTTP(c.Writer, c.Request)
}

// Redirect 重定向
func (c *Context) Redirect(status int, url string) {
	c.Render(status, &render.Redirect{Code: status, Req: c.Request, Location: url})
}

// String 字符串格式化
func (c *Context) String(status int, format string, values ...any) {
	c.Render(status, &render.String{Format: format, Data: values})
}

func (c *Context) Render(status int, render render.Render) {
	err := render.Render(c.Writer)
	if status != http.StatusOK {
		c.Writer.WriteHeader(status)
	}
	if err != nil {
		log.Println(err)
		return
	}
}
