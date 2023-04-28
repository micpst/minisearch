package store

import (
	"github.com/micpst/minisearch/pkg/lib"
	"github.com/micpst/minisearch/pkg/radix"
)

type FindParams struct {
	Tokens    []string
	BoolMode  Mode
	Exact     bool
	Tolerance int
	Relevance BM25Params
	DocsCount int
}

type IndexParams struct {
	Id        string
	Tokens    []string
	DocsCount int
}

type Index struct {
	data             *radix.Trie
	avgFieldLength   float64
	fieldLengths     map[string]int
	tokenOccurrences map[string]int
}

func NewIndex() *Index {
	return &Index{
		data:             radix.New(),
		fieldLengths:     make(map[string]int),
		tokenOccurrences: make(map[string]int),
	}
}

func (idx *Index) Insert(params *IndexParams) {
	tokensCount := lib.Count(params.Tokens)

	for token, count := range tokensCount {
		tokenFrequency := float64(count) / float64(len(params.Tokens))
		idx.data.Insert(&radix.InsertParams{
			Id:            params.Id,
			Word:          token,
			TermFrequency: tokenFrequency,
		})
		idx.tokenOccurrences[token]++
	}

	idx.avgFieldLength = (idx.avgFieldLength*float64(params.DocsCount-1) + float64(len(params.Tokens))) / float64(params.DocsCount)
	idx.fieldLengths[params.Id] = len(params.Tokens)
}

func (idx *Index) Delete(params *IndexParams) {
	for _, token := range params.Tokens {
		idx.data.Delete(&radix.DeleteParams{
			Id:   params.Id,
			Word: token,
		})
		idx.tokenOccurrences[token]--
		if idx.tokenOccurrences[token] == 0 {
			delete(idx.tokenOccurrences, token)
		}
	}

	idx.avgFieldLength = (idx.avgFieldLength*float64(params.DocsCount) - float64(len(params.Tokens))) / float64(params.DocsCount-1)
	delete(idx.fieldLengths, params.Id)
}

func (idx *Index) Find(params *FindParams) map[string]float64 {
	idScores := make(map[string]float64)
	idTokensCount := make(map[string]int)

	for _, token := range params.Tokens {
		infos := idx.data.Find(&radix.FindParams{
			Term:      token,
			Tolerance: params.Tolerance,
			Exact:     params.Exact,
		})
		for _, info := range infos {
			idScores[info.Id] += lib.BM25(
				info.TermFrequency,
				idx.tokenOccurrences[token],
				idx.fieldLengths[info.Id],
				idx.avgFieldLength,
				params.DocsCount,
				params.Relevance.K,
				params.Relevance.B,
				params.Relevance.D,
			)
			idTokensCount[info.Id]++
		}
	}

	for id, tokensCount := range idTokensCount {
		if params.BoolMode == AND && tokensCount != len(params.Tokens) {
			delete(idScores, id)
		}
	}

	return idScores
}
