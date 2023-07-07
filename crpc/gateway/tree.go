package gateway

import (
	"strings"
)

type TreeNode struct {
	Name      string
	Child     []*TreeNode
	RoutePath string // 完整路径
	GwName    string // 完整路径
	IsEnd     bool   // 是否为伪节点
}

// 以递归的方式添加path
func putDfs(splits []string, route string, index int, root *TreeNode, gwName string) {
	// 若index超过splits说明所有节点都已经添加完毕
	if index >= len(splits) {
		return
	}
	for _, node := range root.Child {
		if node.Name == splits[index] {
			route += "/" + splits[index]
			putDfs(splits, route, index+1, node, gwName) // 递归查找已有的路由
			return
		}
	}
	addNode := &TreeNode{Name: splits[index], GwName: gwName}
	route += "/" + splits[index]
	addNode.RoutePath = route
	if index == len(splits)-1 {
		addNode.IsEnd = true
	}
	putDfs(splits, route, index+1, addNode, gwName)
	root.Child = append(root.Child, addNode)
}

// Put 添加路径
func (t *TreeNode) Put(path string, gwName string) {
	path = strings.Trim(path, "/")
	splits := strings.Split(path, "/")
	// 若i == 0第一个字符可能是空格，因此需要忽略
	if len(splits) >= 1 {
		putDfs(splits, "", 0, t, gwName)
	}
}

// 以递归的方式查找路径节点
func getDfs(splits []string, index int, root *TreeNode) (*TreeNode, bool) {
	if index >= len(splits) {
		return nil, false
	}
	for _, node := range root.Child {
		if node.Name == "**" {
			return node, true
		}
		if node.Name == splits[index] ||
			node.Name == "*" ||
			strings.Contains(node.Name, ":") {
			if index+1 == len(splits) {
				return node, true
			}
			return getDfs(splits, index+1, node)
		}
	}
	return nil, false
}

// Get 通过path获取节点
func (t *TreeNode) Get(path string) *TreeNode {
	path = strings.Trim(path, "/")
	splits := strings.Split(path, "/")
	node, _ := getDfs(splits, 0, t)
	return node
}
