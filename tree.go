package inu

import (
	"net/http"
	"regexp"
	"strings"
)

type (
	Tree struct {
		root *Node
	}

	Node struct {
		key           string
		value         *NodeValue
		children      []*Node
		regexChildren []*Node
		regex         RegexInfo
	}

	RegexInfo struct {
		name  string
		regex *regexp.Regexp
	}

	NodeValue struct {
		handle    http.HandlerFunc
		pathParam map[string]string
		value     string
	}
)

func NewNode(key string) *Node {
	return &Node{
		key:      key,
		children: []*Node{},
	}
}

func NewTree() *Tree {
	return &Tree{
		root: NewNode("/"),
	}
}

func (t *Tree) Add(pattern string, value *NodeValue) {
	var currentNode = t.root
	if pattern != currentNode.key {
		pattern = strings.TrimPrefix(pattern, "/")
		nodKeys := strings.Split(pattern, "/")
	l:
		for _, key := range nodKeys {
			regex := fmtRegex(key)
			if regex == nil {
				for _, node := range currentNode.children {
					if node.key == key {
						currentNode = node
						continue l
					}
				}
				node := NewNode(key)
				currentNode.children = append(currentNode.children, node)
				currentNode = node
			} else {
				for _, node := range currentNode.regexChildren {
					if node.key == key {
						currentNode = node
						continue l
					}
				}
				node := NewNode(key)
				node.regex = *regex
				currentNode.regexChildren = append(currentNode.regexChildren, node)
				currentNode = node
			}
		}
	}
	if currentNode.value != nil {
		panic("this url has been defined!")
	}
	currentNode.value = value
}

func (t *Tree) Find(pattern string, suffix bool) (*Node, map[string]string) {
	var currentNode = t.root
	pathParam := make(map[string]string)
	if pattern != currentNode.key {
		pattern = strings.TrimPrefix(pattern, "/")
		if suffix {
			pattern = strings.TrimSuffix(pattern, "/")
		}
		nodKeys := strings.Split(pattern, "/")
		if nod, param := currentNode.Find(nodKeys, pathParam); nod != nil && nod.value != nil {
			return nod, param
		} else {
			return nil, param
		}
	}
	return currentNode, pathParam
}

func (n *Node) Find(nodKeys []string, pathParam map[string]string) (*Node, map[string]string) {
	if len(nodKeys) == 0 {
		return n, pathParam
	}
	key := nodKeys[0]
	for _, node := range n.children {
		if node.key == key {
			return node.Find(nodKeys[1:], pathParam)
		}
	}
	for _, node := range n.regexChildren {
		if str := matchRegexNode(*node, key); str != "" {
			if len(nodKeys) == 1 {
				pathParam[node.regex.name] = key
				return node, pathParam
			}
			nd, param := node.Find(nodKeys[1:], pathParam)
			if nd != nil {
				param[node.regex.name] = str
				return nd, param
			}
		}
	}
	return nil, pathParam
}

func matchRegexNode(node Node, key string) string {
	if node.regex.regex == nil {
		return key
	} else {
		return string(node.regex.regex.Find([]byte(key)))
	}
}

func fmtRegex(str string) *RegexInfo {
	if !strings.HasPrefix(str, "{") || !strings.HasSuffix(str, "}") {
		return nil
	}
	str = strings.TrimSuffix(strings.TrimPrefix(str, "{"), "}")
	spIdx := strings.IndexAny(str, ":")
	switch spIdx {
	case -1:
		return &RegexInfo{name: str}
	case len(str) - 1:
		return &RegexInfo{name: str[:len(str)-1]}
	default:
		reg := strings.Split(str, ":")
		if r, err := regexp.Compile(reg[1]); err != nil {
			panic("url regexp err")
		} else {
			return &RegexInfo{name: reg[0], regex: r}
		}
	}
}
