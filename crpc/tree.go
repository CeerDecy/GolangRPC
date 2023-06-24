package crpc

import (
	"strings"
)

type treeNode struct {
	name      string
	child     []*treeNode
	routePath string // 完整路径
	isEnd     bool   // 是否为伪节点
}

// 以递归的方式添加path
func putDfs(splits []string, route string, index int, root *treeNode) {
	// 若index超过splits说明所有节点都已经添加完毕
	if index >= len(splits) {
		return
	}
	for _, node := range root.child {
		if node.name == splits[index] {
			route += "/" + splits[index]
			putDfs(splits, route, index+1, node) // 递归查找已有的路由
			return
		}
	}
	addNode := &treeNode{name: splits[index]}
	route += "/" + splits[index]
	addNode.routePath = route
	if index == len(splits)-1 {
		addNode.isEnd = true
	}
	putDfs(splits, route, index+1, addNode)
	root.child = append(root.child, addNode)
}

// Put 添加路径
func (t *treeNode) Put(path string) {
	path = strings.Trim(path, "/")
	splits := strings.Split(path, "/")
	// 若i == 0第一个字符可能是空格，因此需要忽略
	if len(splits) >= 1 {
		putDfs(splits, "", 0, t)
	}
}

// 以递归的方式查找路径节点
func getDfs(splits []string, index int, root *treeNode) (*treeNode, bool) {
	if index >= len(splits) {
		return nil, false
	}
	if root.name == splits[index] {
		return root, true
	}
	for _, node := range root.child {
		if node.name == "**" {
			return node, true
		}
		if node.name == splits[index] ||
			node.name == "*" ||
			strings.Contains(node.name, ":") {
			if index+1 == len(splits) {
				return node, true
			}
			return getDfs(splits, index+1, node)
		}
	}
	return nil, false
}

// Get 通过path获取节点
func (t *treeNode) Get(path string) *treeNode {
	path = strings.Trim(path, "/")
	splits := strings.Split(path, "/")
	node, _ := getDfs(splits, 0, t)
	return node
}
