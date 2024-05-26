package radix

import (
	"maps"

	"github.com/micpst/minisearch/pkg/lib"
)

type node[K Key, V Value] struct {
	subword  []rune
	children map[rune]*node[K, V]
	data     map[K]V
}

func newNode[K Key, V Value](subword []rune) *node[K, V] {
	return &node[K, V]{
		subword:  subword,
		children: make(map[rune]*node[K, V]),
		data:     make(map[K]V),
	}
}

func (n *node[K, V]) addChild(child *node[K, V]) {
	if len(child.subword) > 0 {
		n.children[child.subword[0]] = child
	}
}

func (n *node[K, V]) removeChild(child *node[K, V]) {
	if len(child.subword) > 0 {
		delete(n.children, child.subword[0])
	}
}

func (n *node[K, V]) addData(id K, data V) {
	n.data[id] = data
}

func (n *node[K, V]) removeData(id K) {
	delete(n.data, id)
}

func (n *node[K, V]) findData(word []rune, term []rune, tolerance int, exact bool) map[K]V {
	results := make(map[K]V)
	stack := [][2]interface{}{{n, word}}

	for len(stack) > 0 {
		currNode, currWord := stack[len(stack)-1][0].(*node[K, V]), stack[len(stack)-1][1].([]rune)
		stack = stack[:len(stack)-1]

		if _, eq := lib.CommonPrefix(currWord, term); !eq && exact {
			break
		}

		if tolerance > 0 {
			if _, isBounded := lib.BoundedLevenshtein(currWord, term, tolerance); isBounded {
				maps.Copy(results, currNode.data)
			}
		} else {
			maps.Copy(results, currNode.data)
		}

		for _, child := range currNode.children {
			stack = append(stack, [2]interface{}{child, append(currWord, child.subword...)})
		}
	}

	return results
}

func (n *node[K, V]) mergeNode(other *node[K, V]) {
	n.subword = append(n.subword, other.subword...)
	n.data = other.data
	n.children = other.children
}
