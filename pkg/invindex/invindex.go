package invindex

type RecordInfo struct {
	TermFrequency float64
}

type InvIndex map[string]map[string]RecordInfo

func (idx InvIndex) Add(id string, term string, termFrequency float64) {
	if _, ok := idx[term]; !ok {
		idx[term] = make(map[string]RecordInfo, 1)
	}

	idx[term][id] = RecordInfo{
		TermFrequency: termFrequency,
	}
}

func (idx InvIndex) Remove(id string, term string) {
	if _, ok := idx[term]; ok {
		delete(idx[term], id)

		if len(idx[term]) == 0 {
			delete(idx, term)
		}
	}
}

func (idx InvIndex) Find(term string) map[string]RecordInfo {
	infos, ok := idx[term]
	if !ok {
		return make(map[string]RecordInfo)
	}

	return infos
}
