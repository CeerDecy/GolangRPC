package crpc

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	root := &treeNode{name: "/"}
	root.Put("/abc/test")
	root.Put("/abc/ttt")
	fmt.Println(root.child)
}
