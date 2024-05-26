package store

import (
	"fmt"
	"reflect"

	"github.com/micpst/minisearch/pkg/lib"
	"github.com/micpst/minisearch/pkg/radix"
	"github.com/micpst/minisearch/pkg/tokenizer"
)

type recordId comparable

type recordInfo struct {
	termFrequency float64
}

type findParams struct {
	term      string
	property  string
	exact     bool
	tolerance int
	relevance BM25Params
	docsCount int
}

type indexParams[K recordId, S Schema] struct {
	id              K
	document        S
	docsCount       int
	language        tokenizer.Language
	tokenizerConfig *tokenizer.Config
}

type index[K recordId, S Schema] struct {
	indexes              map[string]*radix.Trie[K, recordInfo]
	searchableProperties []string
	avgFieldLength       map[string]float64
	fieldLengths         map[string]map[K]int
	tokenOccurrences     map[string]map[string]int
}

func newIndex[K recordId, S Schema]() *index[K, S] {
	idx := &index[K, S]{
		indexes:              make(map[string]*radix.Trie[K, recordInfo]),
		searchableProperties: make([]string, 0),
		avgFieldLength:       make(map[string]float64),
		fieldLengths:         make(map[string]map[K]int),
		tokenOccurrences:     make(map[string]map[string]int),
	}
	idx.build()
	return idx
}

func (idx *index[K, S]) build() {
	var s S
	for key, value := range flattenSchema(s) {
		switch value.(type) {
		case string:
			idx.indexes[key] = radix.New[K, recordInfo]()
			idx.fieldLengths[key] = make(map[K]int)
			idx.tokenOccurrences[key] = make(map[string]int)
			idx.searchableProperties = append(idx.searchableProperties, key)
		default:
			continue
		}
	}
}

func (idx *index[K, S]) insert(params *indexParams[K, S]) {
	document := flattenSchema(params.document)

	for propName, index := range idx.indexes {
		tokens, _ := tokenizer.Tokenize(&tokenizer.TokenizeParams{
			Text:            document[propName].(string),
			Language:        params.language,
			AllowDuplicates: true,
		}, params.tokenizerConfig)

		allTokensCount := float64(len(tokens))
		tokensCount := lib.Count(tokens)

		for token, count := range tokensCount {
			tokenFrequency := float64(count) / allTokensCount
			index.Insert(&radix.InsertParams[K, recordInfo]{
				Id:   params.id,
				Word: token,
				Data: recordInfo{termFrequency: tokenFrequency},
			})
			idx.tokenOccurrences[propName][token]++
		}

		idx.avgFieldLength[propName] = (idx.avgFieldLength[propName]*float64(params.docsCount-1) + allTokensCount) / float64(params.docsCount)
		idx.fieldLengths[propName][params.id] = int(allTokensCount)
	}
}

func (idx *index[K, S]) delete(params *indexParams[K, S]) {
	document := flattenSchema(params.document)

	for propName, index := range idx.indexes {
		tokens, _ := tokenizer.Tokenize(&tokenizer.TokenizeParams{
			Text:            document[propName].(string),
			Language:        params.language,
			AllowDuplicates: false,
		}, params.tokenizerConfig)

		for _, token := range tokens {
			index.Delete(&radix.DeleteParams[K]{
				Id:   params.id,
				Word: token,
			})
			idx.tokenOccurrences[propName][token]--
			if idx.tokenOccurrences[propName][token] == 0 {
				delete(idx.tokenOccurrences[propName], token)
			}
		}

		idx.avgFieldLength[propName] = (idx.avgFieldLength[propName]*float64(params.docsCount) - float64(len(tokens))) / float64(params.docsCount-1)
		delete(idx.fieldLengths[propName], params.id)
	}
}

func (idx *index[K, S]) find(params *findParams) map[K]float64 {
	idScores := make(map[K]float64)

	if index, ok := idx.indexes[params.property]; ok {
		records := index.Find(&radix.FindParams{
			Term:      params.term,
			Tolerance: params.tolerance,
			Exact:     params.exact,
		})
		for id, data := range records {
			idScores[id] = lib.BM25(
				data.termFrequency,
				idx.tokenOccurrences[params.property][params.term],
				idx.fieldLengths[params.property][id],
				idx.avgFieldLength[params.property],
				params.docsCount,
				params.relevance.K,
				params.relevance.B,
				params.relevance.D,
			)
		}
	}

	return idScores
}

func flattenSchema(obj any, prefix ...string) map[string]any {
	m := make(map[string]any)
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	fields := reflect.VisibleFields(t)

	for i, field := range fields {
		if propName, ok := field.Tag.Lookup("index"); ok {
			if len(prefix) == 1 {
				propName = fmt.Sprintf("%s.%s", prefix[0], propName)
			}

			switch field.Type.Kind() {
			case reflect.Struct:
				for key, value := range flattenSchema(v.Field(i).Interface(), propName) {
					m[key] = value
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				m[propName] = v.Field(i).Int()
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				m[propName] = v.Field(i).Uint()
			default:
				m[propName] = v.Field(i).String()
			}
		}
	}

	return m
}
