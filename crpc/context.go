package crpc

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github/CeerDecy/RpcFrameWork/crpc/utils"
	"html/template"
	"log"
	"net/http"
	"net/url"
)

type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
	engine  *Engine
}

// HTML 返回HTML文本
func (c *Context) HTML(status int, html string) {
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(status)
	_, err := c.Writer.Write([]byte(html))
	if err != nil {
		log.Println(err)
		return
	}
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
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := c.engine.HTMLRender.Template.ExecuteTemplate(c.Writer, name, data)
	if err != nil {
		log.Println(err)
	}
}

// JSON 返回JSON数据
func (c *Context) JSON(state int, data any) error {
	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Writer.WriteHeader(state)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = c.Writer.Write(jsonData)
	return err
}

// XML 返回XML数据
func (c *Context) XML(state int, data any) error {
	c.Writer.Header().Set("Content-Type", "application/xml; charset=utf-8")
	c.Writer.WriteHeader(state)
	err := xml.NewEncoder(c.Writer).Encode(data)
	return err
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
