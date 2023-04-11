package radix

import (
	"github.com/micpst/minisearch/pkg/lib"
)

type InsertParams struct {
	Id            string
	Word          string
	TermFrequency float64
}

type DeleteParams struct {
	Id   string
	Word string
}

type FindParams struct {
	Term string
}

type Trie struct {
	root   *node
	length int
}

func New() *Trie {
	return &Trie{root: newNode(nil)}
}

func (t *Trie) Len() int {
	return t.length
}

func (t *Trie) Insert(params *InsertParams) {
	word := []rune(params.Word)
	newInfo := RecordInfo{
		Id:            params.Id,
		TermFrequency: params.TermFrequency,
	}

	currentNode := t.root
	for i := 0; i < len(word); {
		wordAtIndex := word[i:]

		if currentChild, ok := currentNode.children[wordAtIndex[0]]; ok {
			commonPrefix := lib.CommonPrefix(currentChild.subword, wordAtIndex)
			commonPrefixLength := len(commonPrefix)
			subwordLength := len(currentChild.subword)
			wordLength := len(wordAtIndex)

			// the wordAtIndex matches exactly with an existing child node
			if commonPrefixLength == wordLength && commonPrefixLength == subwordLength {
				currentChild.addRecordInfo(newInfo)
				return
			}

			// the wordAtIndex is completely contained in the child node subword
			if commonPrefixLength == wordLength && commonPrefixLength < subwordLength {
				n := newNode(wordAtIndex)
				n.addRecordInfo(newInfo)

				currentChild.subword = currentChild.subword[commonPrefixLength:]
				n.addChild(currentChild)
				currentNode.addChild(n)

				t.length++
				return
			}

			// the wordAtIndex is partially contained in the child node subword
			if commonPrefixLength < wordLength && commonPrefixLength < subwordLength {
				n := newNode(wordAtIndex[commonPrefixLength:])
				n.addRecordInfo(newInfo)

				inBetweenNode := newNode(wordAtIndex[:commonPrefixLength])
				currentNode.addChild(inBetweenNode)

				currentChild.subword = currentChild.subword[commonPrefixLength:]
				inBetweenNode.addChild(currentChild)
				inBetweenNode.addChild(n)

				t.length++
				return
			}

			// skip to the next divergent character
			i += subwordLength

			// navigate in the child node
			currentNode = currentChild
		} else {
			// if the node for the current character doesn't exist create a new child node
			n := newNode(wordAtIndex)
			n.addRecordInfo(newInfo)

			currentNode.addChild(n)
			t.length++
			return
		}
	}
}

func (t *Trie) Delete(params *DeleteParams) {
	if params.Word == "" {
		return
	}

	word := []rune(params.Word)
	currentNode := t.root

	for i := 0; i < len(word); {
		char := word[i]
		wordAtIndex := word[i:]

		if currentChild, ok := currentNode.children[char]; ok {
			commonPrefix := lib.CommonPrefix(currentChild.subword, wordAtIndex)
			commonPrefixLength := len(commonPrefix)
			subwordLength := len(currentChild.subword)
			wordLength := len(wordAtIndex)

			// the wordAtIndex matches exactly with an existing child node
			if commonPrefixLength == wordLength && commonPrefixLength == subwordLength {
				currentChild.removeRecordInfo(params.Id)

				if len(currentChild.infos) == 0 {
					switch len(currentChild.children) {
					case 0:
						// if the node to be deleted has no children, delete it
						delete(currentNode.children, char)
						t.length--
					case 1:
						// if the node to be deleted has one child, promote it to the parent node
						for _, child := range currentChild.children {
							currentChild.subword = append(currentChild.subword, child.subword...)
							currentChild.infos = child.infos
							currentChild.children = child.children
						}
						t.length--
					}
				}
				return
			}

			// skip to the next divergent character
			i += subwordLength

			// navigate in the child node
			currentNode = currentChild
		} else {
			// if the node for the current character doesn't exist abort the deletion
			return
		}
	}
}

func (t *Trie) Find(params *FindParams) RecordInfos {
	word := []rune(params.Term)
	currentNode := t.root

	for i := 0; i < len(word); {
		char := word[i]
		wordAtIndex := word[i:]

		if currentChild, ok := currentNode.children[char]; ok {
			commonPrefix := lib.CommonPrefix(currentChild.subword, wordAtIndex)
			commonPrefixLength := len(commonPrefix)
			subwordLength := len(currentChild.subword)
			wordLength := len(wordAtIndex)

			// the wordAtIndex doesn't match exactly with an existing child node
			if commonPrefixLength != wordLength && commonPrefixLength != subwordLength {
				return RecordInfos{}
			}

			// skip to the next divergent character
			i += subwordLength

			// navigate in the child node
			currentNode = currentChild
		} else {
			// if the node for the current character doesn't exist abort the deletion
			return RecordInfos{}
		}
	}

	return currentNode.findAllRecordInfos()
}
