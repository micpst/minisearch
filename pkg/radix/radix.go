package radix

import (
	"github.com/micpst/minisearch/pkg/lib"
)

type Key comparable
type Value any

type InsertParams[K Key, V Value] struct {
	Id   K
	Word string
	Data V
}

type DeleteParams[K Key] struct {
	Id   K
	Word string
}

type FindParams struct {
	Term      string
	Tolerance int
	Exact     bool
}

type FindResult[V Value] struct {
	Id   string
	Data V
}

type Trie[K Key, V Value] struct {
	root   *node[K, V]
	length int
}

func New[K Key, V Value]() *Trie[K, V] {
	return &Trie[K, V]{root: newNode[K, V](nil)}
}

func (t *Trie[K, V]) Len() int {
	return t.length
}

func (t *Trie[K, V]) Insert(params *InsertParams[K, V]) {
	word := []rune(params.Word)
	currNode := t.root

	for i := 0; i < len(word); {
		wordAtIndex := word[i:]

		if currChild, ok := currNode.children[wordAtIndex[0]]; ok {
			commonPrefix, _ := lib.CommonPrefix(currChild.subword, wordAtIndex)
			commonPrefixLength := len(commonPrefix)
			subwordLength := len(currChild.subword)
			wordLength := len(wordAtIndex)

			// the wordAtIndex matches exactly with an existing child node
			if commonPrefixLength == wordLength && commonPrefixLength == subwordLength {
				currChild.addData(params.Id, params.Data)
				return
			}

			// the wordAtIndex is completely contained in the child node subword
			if commonPrefixLength == wordLength && commonPrefixLength < subwordLength {
				n := newNode[K, V](wordAtIndex)
				n.addData(params.Id, params.Data)

				currChild.subword = currChild.subword[commonPrefixLength:]
				n.addChild(currChild)
				currNode.addChild(n)

				t.length++
				return
			}

			// the wordAtIndex is partially contained in the child node subword
			if commonPrefixLength < wordLength && commonPrefixLength < subwordLength {
				n := newNode[K, V](wordAtIndex[commonPrefixLength:])
				n.addData(params.Id, params.Data)

				inBetweenNode := newNode[K, V](wordAtIndex[:commonPrefixLength])
				currNode.addChild(inBetweenNode)

				currChild.subword = currChild.subword[commonPrefixLength:]
				inBetweenNode.addChild(currChild)
				inBetweenNode.addChild(n)

				t.length++
				return
			}

			// skip to the next divergent character
			i += subwordLength

			// navigate in the child node
			currNode = currChild
		} else {
			// if the node for the curr character doesn't exist create a new child node
			n := newNode[K, V](wordAtIndex)
			n.addData(params.Id, params.Data)

			currNode.addChild(n)
			t.length++
			return
		}
	}
}

func (t *Trie[K, V]) Delete(params *DeleteParams[K]) {
	word := []rune(params.Word)
	currNode := t.root

	for i := 0; i < len(word); {
		char := word[i]
		wordAtIndex := word[i:]

		if currChild, ok := currNode.children[char]; ok {
			if _, eq := lib.CommonPrefix(currChild.subword, wordAtIndex); eq {
				currChild.removeData(params.Id)

				if len(currChild.data) == 0 {
					switch len(currChild.children) {
					case 0:
						// if the node to be deleted has no children, delete it
						currNode.removeChild(currChild)
						t.length--
					case 1:
						// if the node to be deleted has one child, promote it to the parent node
						for _, child := range currChild.children {
							currChild.mergeNode(child)
						}
						t.length--
					}
				}
				return
			}

			// skip to the next divergent character
			i += len(currChild.subword)

			// navigate in the child node
			currNode = currChild
		} else {
			// if the node for the curr character doesn't exist abort the deletion
			return
		}
	}
}

func (t *Trie[K, V]) Find(params *FindParams) map[K]V {
	term := []rune(params.Term)
	currNode := t.root
	currNodeWord := currNode.subword

	for i := 0; i < len(term); {
		char := term[i]
		wordAtIndex := term[i:]

		if currChild, ok := currNode.children[char]; ok {
			commonPrefix, _ := lib.CommonPrefix(currChild.subword, wordAtIndex)
			commonPrefixLength := len(commonPrefix)
			subwordLength := len(currChild.subword)
			wordLength := len(wordAtIndex)

			// if the common prefix length is equal to the node subword length it means they are a match
			// if the common prefix is equal to the term means it is contained in the node
			if commonPrefixLength != wordLength && commonPrefixLength != subwordLength {
				if params.Tolerance > 0 {
					break
				}
				return map[K]V{}
			}

			// skip to the next divergent character
			i += subwordLength

			// navigate in the child node
			currNode = currChild

			// update the current node word
			currNodeWord = append(currNodeWord, currChild.subword...)
		} else {
			// if the node for the curr character doesn't exist abort the deletion
			return map[K]V{}
		}
	}

	return currNode.findData(currNodeWord, term, params.Tolerance, params.Exact)
}
