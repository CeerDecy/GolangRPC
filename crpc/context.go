package crpc

import (
	"log"
	"net/http"
)

type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
}

func (c *Context) HTML(status int, html string) {
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(status)
	_, err := c.Writer.Write([]byte(html))
	if err != nil {
		log.Fatalln(err)
	}
}
