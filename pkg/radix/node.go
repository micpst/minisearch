package radix

import (
	"sort"
)

type RecordInfo struct {
	Id            string
	TermFrequency float64
}

type RecordInfos []RecordInfo

func (r RecordInfos) Len() int { return len(r) }

func (r RecordInfos) Swap(i, j int) { r[i], r[j] = r[j], r[i] }

func (r RecordInfos) Less(i, j int) bool { return r[i].TermFrequency < r[j].TermFrequency }

func (r RecordInfos) Sort() { sort.Sort(r) }

type node struct {
	subword  []rune
	children map[rune]*node
	infos    RecordInfos
}

func newNode(subword []rune) *node {
	return &node{
		subword:  subword,
		children: make(map[rune]*node),
		infos:    make(RecordInfos, 0),
	}
}

func (n *node) addChild(child *node) {
	if len(child.subword) > 0 {
		n.children[child.subword[0]] = child
	}
}

func (n *node) addRecordInfo(info RecordInfo) {
	num := len(n.infos)
	idx := sort.Search(num, func(i int) bool {
		return n.infos[i].Id >= info.Id
	})

	n.infos = append(n.infos, RecordInfo{})
	copy(n.infos[idx+1:], n.infos[idx:])
	n.infos[idx] = info
}

func (n *node) updateRecordInfo(info RecordInfo) bool {
	num := len(n.infos)
	idx := sort.Search(num, func(i int) bool {
		return n.infos[i].Id >= info.Id
	})

	if idx < num && n.infos[idx].Id == info.Id {
		n.infos[idx] = info
		return true
	}
	return false
}

func (n *node) removeRecordInfo(id string) bool {
	num := len(n.infos)
	idx := sort.Search(num, func(i int) bool {
		return n.infos[i].Id >= id
	})

	if idx < num && n.infos[idx].Id == id {
		copy(n.infos[idx:], n.infos[idx+1:])
		n.infos[len(n.infos)-1] = RecordInfo{}
		n.infos = n.infos[:len(n.infos)-1]
		return true
	}
	return false
}

func (n *node) getRecordInfo(id string) *RecordInfo {
	num := len(n.infos)
	idx := sort.Search(num, func(i int) bool {
		return n.infos[i].Id >= id
	})

	if idx < num && n.infos[idx].Id == id {
		return &n.infos[idx]
	}
	return nil
}
