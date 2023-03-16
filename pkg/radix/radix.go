package radix

import (
	"github.com/micpst/minisearch/pkg/lib"
)

type InsertParams struct {
	Id            string
	Word          string
	TermFrequency float64
}

type Trie struct {
	root *node
}

func New() *Trie {
	return &Trie{root: newNode(nil)}
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

			if commonPrefixLength == wordLength {
				if commonPrefixLength == subwordLength {
					// the wordAtIndex matches exactly with an existing child node
					currentChild.addRecordInfo(newInfo)
					return

				} else {
					// the wordAtIndex is completely contained in the child node subword
					n := newNode(wordAtIndex)
					n.addRecordInfo(newInfo)

					currentChild.subword = currentChild.subword[commonPrefixLength:]
					n.addChild(currentChild)
					currentNode.addChild(n)
					return
				}

			} else if commonPrefixLength < subwordLength {
				// the wordAtIndex is partially contained in the child node subword
				n := newNode(wordAtIndex[commonPrefixLength:])
				n.addRecordInfo(newInfo)

				inBetweenNode := newNode(wordAtIndex[:commonPrefixLength])
				currentNode.addChild(inBetweenNode)

				currentChild.subword = currentChild.subword[commonPrefixLength:]
				inBetweenNode.addChild(currentChild)
				inBetweenNode.addChild(n)
				return
			}

			// skip to the next divergent character
			i += subwordLength

			// navigate in the child node
			currentNode = currentChild

		} else {
			// if the node for the current character doesn't exist create new child node
			n := newNode(wordAtIndex)
			n.addRecordInfo(newInfo)

			currentNode.addChild(n)
			return
		}
	}
}
