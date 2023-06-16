package crpc

import "strings"

type treeNode struct {
	name  string
	child []*treeNode
}

// 以递归的方式添加path
func putDfs(splits []string, index int, root *treeNode) {
	if index >= len(splits) {
		return
	}
	for _, node := range root.child {
		if node.name == splits[index] {
			putDfs(splits, index+1, node)
			return
		}
	}
	addNode := &treeNode{name: splits[index]}
	putDfs(splits, index+1, addNode)
	root.child = append(root.child, addNode)
}

// Put 添加路径
func (t *treeNode) Put(path string) {
	splits := strings.Split(path, "/")
	// 若i == 0第一个字符可能是空格，因此需要忽略
	if len(splits) > 1 {
		putDfs(splits, 1, t)
	}
}
