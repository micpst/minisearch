package store

import (
	"fmt"
	"math"
	"reflect"
	"sync"

	"github.com/google/uuid"
	"github.com/micpst/fts-engine/pkg/lib"
)

type SchemaProps any

type Record[Schema SchemaProps] struct {
	Id string
	S  Schema
}

type Mode string

type SearchParams struct {
	Query      string
	Properties []string
	BoolMode   Mode
}

type findParams struct {
	query    string
	boolMode Mode
}

type recordInfo struct {
	freq float64
}

type memIndex struct {
	index map[string]map[string]recordInfo
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

	return Record[Schema]{Id: id, S: doc}, nil
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

	return Record[Schema]{Id: id, S: doc}, nil
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

func (db *MemDB[Schema]) Search(params SearchParams) []Record[Schema] {
	recordsIds := make(map[string]struct{})
	records := make([]Record[Schema], 0)
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
			ids := index.find(findParams{
				query:    params.Query,
				boolMode: params.BoolMode,
			})
			for _, id := range ids {
				recordsIds[id] = struct{}{}
			}
		}
	}

	for id := range recordsIds {
		if doc, ok := db.docs[id]; ok {
			records = append(records, Record[Schema]{Id: id, S: doc})
		}
	}

	return records
}

func newIndex() *memIndex {
	return &memIndex{
		index: make(map[string]map[string]recordInfo),
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
			freq: float64(count) / float64(len(tokens)),
		}
		idx.index[token][id] = info
	}
}

func (idx *memIndex) remove(id string, text string) {
	tokens := lib.Tokenize(text)

	for _, token := range tokens {
		if _, ok := idx.index[token]; ok {
			delete(idx.index[token], id)
		}
	}
}

func (idx *memIndex) find(params findParams) []string {
	resultIds := make([]string, 0)
	recordIds := make(map[string]int)

	tokens := lib.Tokenize(params.query)

	for _, token := range tokens {
		if infos, ok := idx.index[token]; ok {
			for id := range infos {
				recordIds[id] += 1
			}
		}
	}

	for id, tokensCount := range recordIds {
		if (params.boolMode == AND && tokensCount == len(tokens)) ||
			params.boolMode == OR {
			resultIds = append(resultIds, id)
		}
	}

	return resultIds
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
