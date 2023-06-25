package crpc

import (
	"html/template"
	"log"
	"net/http"
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
