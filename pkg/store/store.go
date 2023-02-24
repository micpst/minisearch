package store

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/micpst/fts-engine/pkg/lib"
)

type SchemaProps any

type Record[Schema SchemaProps] struct {
	Id   string
	Data Schema
}

type Mode string

type SearchParams struct {
	Query      string
	Properties []string
	BoolMode   Mode
}

type SearchResult[Schema SchemaProps] struct {
	Id    string
	Data  Schema
	Score float64
}

type SearchResults[Schema SchemaProps] []SearchResult[Schema]

type findParams struct {
	query    string
	boolMode Mode
}

type recordInfo struct {
	termFrequency float64
}

type memIndex struct {
	index       map[string]map[string]recordInfo
	occurrences map[string]int
	docsCount   int
}

type MemDB[Schema SchemaProps] struct {
	mutex   sync.RWMutex
	docs    map[string]Schema
	indexes map[string]*memIndex
}

const (
	AND Mode = "AND"
	OR  Mode = "OR"
)

const WILDCARD = "*"

func New[Schema SchemaProps]() *MemDB[Schema] {
	db := &MemDB[Schema]{
		docs:    make(map[string]Schema),
		indexes: make(map[string]*memIndex),
	}
	db.buildIndexes()
	return db
}

func (db *MemDB[Schema]) buildIndexes() {
	var s Schema
	for key := range flattenSchema(s) {
		db.indexes[key] = newIndex()
	}
}

func (db *MemDB[Schema]) Insert(doc Schema) (Record[Schema], error) {
	id := uuid.NewString()
	docMap := flattenSchema(doc)

	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, ok := db.docs[id]; ok {
		return Record[Schema]{}, fmt.Errorf("document id already exists")
	}

	db.docs[id] = doc

	for propName, index := range db.indexes {
		index.add(id, docMap[propName])
	}

	return Record[Schema]{Id: id, Data: doc}, nil
}

func (db *MemDB[Schema]) InsertBatch(docs []Schema, batchSize int) []error {
	batchCount := int(math.Ceil(float64(len(docs)) / float64(batchSize)))
	docsChan := make(chan Schema)
	errsChan := make(chan error)

	var wg sync.WaitGroup
	wg.Add(batchCount)

	go func() {
		for _, doc := range docs {
			docsChan <- doc
		}
		close(docsChan)
	}()

	for i := 0; i < batchCount; i++ {
		go func() {
			defer wg.Done()
			for doc := range docsChan {
				if _, err := db.Insert(doc); err != nil {
					errsChan <- err
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(errsChan)
	}()

	errs := make([]error, 0)
	for err := range errsChan {
		errs = append(errs, err)
	}

	return errs
}

func (db *MemDB[Schema]) Update(id string, doc Schema) (Record[Schema], error) {
	docMap := flattenSchema(doc)

	db.mutex.Lock()
	defer db.mutex.Unlock()

	prevDoc, ok := db.docs[id]
	if !ok {
		return Record[Schema]{}, fmt.Errorf("document not found")
	}
	prevDocMap := flattenSchema(prevDoc)

	for propName, index := range db.indexes {
		index.remove(id, prevDocMap[propName])
		index.add(id, docMap[propName])
	}

	db.docs[id] = doc

	return Record[Schema]{Id: id, Data: doc}, nil
}

func (db *MemDB[Schema]) Delete(id string) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	doc, ok := db.docs[id]
	if !ok {
		return fmt.Errorf("document not found")
	}
	docMap := flattenSchema(doc)

	for propName, index := range db.indexes {
		index.remove(id, docMap[propName])
	}

	delete(db.docs, id)

	return nil
}

func (db *MemDB[Schema]) Search(params SearchParams) []SearchResult[Schema] {
	resultIdScores := make(map[string]float64)
	results := make(SearchResults[Schema], 0)
	props := params.Properties

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	if len(props) == 1 && props[0] == WILDCARD {
		props = make([]string, 0)
		var s Schema
		for key := range flattenSchema(s) {
			props = append(props, key)
		}
	}

	for _, prop := range props {
		if index, ok := db.indexes[prop]; ok {
			idScores := index.find(findParams{
				query:    params.Query,
				boolMode: params.BoolMode,
			})
			for id, score := range idScores {
				resultIdScores[id] += score
			}
		}
	}

	for id, score := range resultIdScores {
		if doc, ok := db.docs[id]; ok {
			results = append(results, SearchResult[Schema]{
				Id:    id,
				Data:  doc,
				Score: score,
			})
		}
	}

	sort.Sort(results)

	return results
}

func (r SearchResults[Schema]) Len() int {
	return len(r)
}

func (r SearchResults[Schema]) Less(i, j int) bool {
	return r[i].Score > r[j].Score
}

func (r SearchResults[Schema]) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func newIndex() *memIndex {
	return &memIndex{
		index:       make(map[string]map[string]recordInfo),
		occurrences: make(map[string]int),
	}
}

func (idx *memIndex) add(id string, text string) {
	tokens := lib.Tokenize(text)
	tokensCount := lib.Count(tokens)

	for token, count := range tokensCount {
		if _, ok := idx.index[token]; !ok {
			idx.index[token] = make(map[string]recordInfo)
		}

		info := recordInfo{
			termFrequency: float64(count) / float64(len(tokens)),
		}
		idx.index[token][id] = info
		idx.occurrences[token]++
	}

	idx.docsCount++
}

func (idx *memIndex) remove(id string, text string) {
	tokens := lib.Tokenize(text)

	for _, token := range tokens {
		if _, ok := idx.index[token]; ok {
			idx.occurrences[token]--

			if idx.occurrences[token] <= 0 {
				delete(idx.index, token)
				delete(idx.occurrences, token)
			} else {
				delete(idx.index[token], id)
			}
		}
	}

	idx.docsCount--
}

func (idx *memIndex) find(params findParams) map[string]float64 {
	idScores := make(map[string]float64)
	idTokensCount := make(map[string]int)

	tokens := lib.Tokenize(params.query)

	for _, token := range tokens {
		if infos, ok := idx.index[token]; ok {
			for id, info := range infos {
				idScores[id] += lib.TfIdf(info.termFrequency, idx.occurrences[token], idx.docsCount)
				idTokensCount[id]++
			}
		}
	}

	for id, tokensCount := range idTokensCount {
		if params.boolMode == AND && tokensCount != len(tokens) {
			delete(idScores, id)
		}
	}

	return idScores
}

func flattenSchema(obj any, prefix ...string) map[string]string {
	m := make(map[string]string)
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	fields := reflect.VisibleFields(t)

	for i, field := range fields {
		if propName, ok := field.Tag.Lookup("index"); ok {
			if len(prefix) == 1 {
				propName = fmt.Sprintf("%s.%s", prefix[0], propName)
			}

			if field.Type.Kind() == reflect.Struct {
				for key, value := range flattenSchema(v.Field(i).Interface(), propName) {
					m[key] = value
				}
			} else {
				m[propName] = v.Field(i).String()
			}
		}
	}

	return m
}
