package crpc

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	//fmt.Println()
	root := &treeNode{name: "/"}
	root.Put("/abc/test")
	root.Put("/abc/ttt")
	root.Put("/def/**")
	root.Put("/gh/:id")
	fmt.Println(root.Get("/abc/ttt"))
	fmt.Println(root.Get("/abc/sss"))
	fmt.Println(root.Get("/def/3443/fudn"))
	fmt.Println(root.Get("/def/3443"))
	fmt.Println(root.Get("/gh/3443"))
}
