package crpc

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	//fmt.Println()
	root := &treeNode{name: "/user"}
	//root.Put("/abc/test")
	//root.Put("/abc/ttt")
	//root.Put("/def/**")
	//root.Put("/gh/:id")
	//root.Put("/html")
	root.Put("/imageByName")
	fmt.Println(root.Get("/imageByName"))
	//fmt.Println(root.Get("/abc/ttt"))
	//fmt.Println(root.Get("/abc/sss"))
	//fmt.Println(root.Get("/def/3443/fudn"))
	//fmt.Println(root.Get("/def/3443"))
	//fmt.Println(root.Get("/gh/3443"))
	//fmt.Println(root.Get("/html"))
}
