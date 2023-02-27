package store

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/micpst/minisearch/pkg/invindex"
	"github.com/micpst/minisearch/pkg/lib"
)

type SchemaProps any

type Record[Schema SchemaProps] struct {
	Id   string
	Data Schema
}

type Mode string

const (
	AND Mode = "AND"
	OR  Mode = "OR"
)

type SearchParams struct {
	Query      string
	Properties []string
	BoolMode   Mode
	Offset     int
	Limit      int
}

type SearchResult[Schema SchemaProps] struct {
	Hits  SearchHits[Schema]
	Count int
}

type SearchHit[Schema SchemaProps] struct {
	Id    string
	Data  Schema
	Score float64
}

type SearchHits[Schema SchemaProps] []SearchHit[Schema]

func (r SearchHits[Schema]) Len() int { return len(r) }

func (r SearchHits[Schema]) Swap(i, j int) { r[i], r[j] = r[j], r[i] }

func (r SearchHits[Schema]) Less(i, j int) bool { return r[i].Score > r[j].Score }

type MemDB[Schema SchemaProps] struct {
	mu          sync.RWMutex
	docs        map[string]Schema
	indexes     map[string]invindex.InvIndex
	indexKeys   []string
	occurrences map[string]map[string]int
}

const WILDCARD = "*"

func New[Schema SchemaProps]() *MemDB[Schema] {
	db := &MemDB[Schema]{
		docs:        make(map[string]Schema),
		indexes:     make(map[string]invindex.InvIndex),
		indexKeys:   make([]string, 0),
		occurrences: make(map[string]map[string]int),
	}
	db.buildIndexes()
	return db
}

func (db *MemDB[Schema]) buildIndexes() {
	var s Schema
	for key := range flattenSchema(s) {
		db.indexes[key] = invindex.InvIndex{}
		db.indexKeys = append(db.indexKeys, key)
		db.occurrences[key] = make(map[string]int)
	}
}

func (db *MemDB[Schema]) Insert(doc Schema) (Record[Schema], error) {
	id := uuid.NewString()
	docMap := flattenSchema(doc)

	db.mu.Lock()
	defer db.mu.Unlock()

	if _, ok := db.docs[id]; ok {
		return Record[Schema]{}, fmt.Errorf("document id already exists")
	}

	db.docs[id] = doc

	for propName, index := range db.indexes {
		db.indexDocumentField(index, propName, id, docMap)
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

	db.mu.Lock()
	defer db.mu.Unlock()

	oldDoc, ok := db.docs[id]
	if !ok {
		return Record[Schema]{}, fmt.Errorf("document not found")
	}

	oldDocMap := flattenSchema(oldDoc)
	for propName, index := range db.indexes {
		db.deindexDocumentField(index, propName, id, oldDocMap)
		db.indexDocumentField(index, propName, id, docMap)
	}

	db.docs[id] = doc

	return Record[Schema]{Id: id, Data: doc}, nil
}

func (db *MemDB[Schema]) Delete(id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	doc, ok := db.docs[id]
	if !ok {
		return fmt.Errorf("document not found")
	}

	docMap := flattenSchema(doc)
	for propName, index := range db.indexes {
		db.deindexDocumentField(index, propName, id, docMap)
	}

	delete(db.docs, id)

	return nil
}

func (db *MemDB[Schema]) Search(params SearchParams) SearchResult[Schema] {
	idScores := make(map[string]float64)
	results := make(SearchHits[Schema], 0)

	props := params.Properties
	if len(props) == 1 && props[0] == WILDCARD {
		props = db.indexKeys
	}

	tokens := lib.Tokenize(params.Query)

	db.mu.RLock()
	defer db.mu.RUnlock()

	for _, prop := range props {
		if index, ok := db.indexes[prop]; ok {
			idTokensCount := make(map[string]int)

			for _, token := range tokens {
				idInfos := index.Find(token)
				for id, info := range idInfos {
					idScores[id] += lib.TfIdf(info.TermFrequency, db.occurrences[prop][token], len(db.docs))
					idTokensCount[id]++
				}
			}

			for id, tokensCount := range idTokensCount {
				if params.BoolMode == AND && tokensCount != len(tokens) {
					delete(idScores, id)
				}
			}
		}
	}

	for id, score := range idScores {
		if doc, ok := db.docs[id]; ok {
			results = append(results, SearchHit[Schema]{
				Id:    id,
				Data:  doc,
				Score: score,
			})
		}
	}

	sort.Sort(results)

	start, stop := lib.Paginate(params.Offset, params.Limit, len(results))

	return SearchResult[Schema]{
		Hits:  results[start:stop],
		Count: len(results),
	}
}

func (db *MemDB[Schema]) indexDocumentField(index invindex.InvIndex, propName string, id string, docMap map[string]string) {
	tokens := lib.Tokenize(docMap[propName])
	tokensCount := lib.Count(tokens)

	for token, count := range tokensCount {
		tokenFrequency := float64(count) / float64(len(tokens))
		index.Add(id, token, tokenFrequency)

		db.occurrences[propName][token]++
	}
}

func (db *MemDB[Schema]) deindexDocumentField(index invindex.InvIndex, propName string, id string, docMap map[string]string) {
	tokens := lib.Tokenize(docMap[propName])

	for _, token := range tokens {
		index.Remove(id, token)

		db.occurrences[propName][token]--
		if db.occurrences[propName][token] == 0 {
			delete(db.occurrences[propName], token)
		}
	}
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
