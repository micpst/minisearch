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
	Term      string
	Tolerance int
	Exact     bool
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
				currChild.addRecordInfo(newInfo)
				return
			}

			// the wordAtIndex is completely contained in the child node subword
			if commonPrefixLength == wordLength && commonPrefixLength < subwordLength {
				n := newNode(wordAtIndex)
				n.addRecordInfo(newInfo)

				currChild.subword = currChild.subword[commonPrefixLength:]
				n.addChild(currChild)
				currNode.addChild(n)

				t.length++
				return
			}

			// the wordAtIndex is partially contained in the child node subword
			if commonPrefixLength < wordLength && commonPrefixLength < subwordLength {
				n := newNode(wordAtIndex[commonPrefixLength:])
				n.addRecordInfo(newInfo)

				inBetweenNode := newNode(wordAtIndex[:commonPrefixLength])
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
			n := newNode(wordAtIndex)
			n.addRecordInfo(newInfo)

			currNode.addChild(n)
			t.length++
			return
		}
	}
}

func (t *Trie) Delete(params *DeleteParams) {
	word := []rune(params.Word)
	currNode := t.root

	for i := 0; i < len(word); {
		char := word[i]
		wordAtIndex := word[i:]

		if currChild, ok := currNode.children[char]; ok {
			if _, eq := lib.CommonPrefix(currChild.subword, wordAtIndex); eq {
				currChild.removeRecordInfo(params.Id)

				if len(currChild.infos) == 0 {
					switch len(currChild.children) {
					case 0:
						// if the node to be deleted has no children, delete it
						currNode.removeChild(currChild)
						t.length--
					case 1:
						// if the node to be deleted has one child, promote it to the parent node
						for _, child := range currChild.children {
							mergeNodes(currChild, child)
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

func (t *Trie) Find(params *FindParams) []RecordInfo {
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
				return []RecordInfo{}
			}

			// skip to the next divergent character
			i += subwordLength

			// navigate in the child node
			currNode = currChild

			// update the current node word
			currNodeWord = append(currNodeWord, currChild.subword...)
		} else {
			// if the node for the curr character doesn't exist abort the deletion
			return []RecordInfo{}
		}
	}

	return findAllRecordInfos(currNode, currNodeWord, term, params.Tolerance, params.Exact)
}
