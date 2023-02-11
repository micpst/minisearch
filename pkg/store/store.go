package store

import (
	"fmt"
	"math"
	"reflect"
	"sync"

	"github.com/cornelk/hashmap"
	"github.com/google/uuid"
	"github.com/micpst/fts-engine/pkg/lib"
)

type SchemaProps any

type Record[Schema SchemaProps] struct {
	Id string
	S  Schema
}

type RecordInfo struct {
	freq uint32
}

type Mode string

const (
	AND Mode = "AND"
	OR  Mode = "OR"
)

const WILDCARD = "*"

type SearchParams struct {
	Query      string
	Properties []string
	BoolMode   Mode
}

type FindParams struct {
	Query    string
	BoolMode Mode
}

type MemIndex struct {
	*hashmap.Map[string, *hashmap.Map[string, RecordInfo]]
}

type MemDB[Schema SchemaProps] struct {
	docs    *hashmap.Map[string, Schema]
	indexes *hashmap.Map[string, *MemIndex]
}

func New[Schema SchemaProps]() *MemDB[Schema] {
	db := &MemDB[Schema]{
		docs:    hashmap.New[string, Schema](),
		indexes: hashmap.New[string, *MemIndex](),
	}
	db.buildIndexes()
	return db
}

func (db *MemDB[Schema]) buildIndexes() {
	var s Schema
	for key := range schemaToFlatMap(s) {
		db.indexes.Set(key, NewIndex())
	}
}

func (db *MemDB[Schema]) Insert(doc Schema) (Record[Schema], error) {
	id := uuid.NewString()
	if ok := db.docs.Insert(id, doc); !ok {
		return Record[Schema]{}, fmt.Errorf("document cannot be created")
	}

	docMap := schemaToFlatMap(doc)

	db.indexes.Range(func(propName string, index *MemIndex) bool {
		index.Add(id, docMap[propName])
		return true
	})

	return Record[Schema]{Id: id, S: doc}, nil
}

func (db *MemDB[Schema]) InsertBatch(docs []Schema, batchSize int) []error {
	in := make(chan Schema)
	out := make(chan error)
	n := int(math.Ceil(float64(len(docs)) / float64(batchSize)))

	var wg sync.WaitGroup
	wg.Add(n)

	go func() {
		for _, d := range docs {
			in <- d
		}
		close(in)
	}()

	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			for d := range in {
				if _, err := db.Insert(d); err != nil {
					out <- err
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	errs := make([]error, 0)
	for err := range out {
		errs = append(errs, err)
	}

	return errs
}

func (db *MemDB[Schema]) Update(id string, doc Schema) (Record[Schema], error) {
	prevDoc, ok := db.docs.Get(id)
	if !ok {
		return Record[Schema]{}, fmt.Errorf("document not found")
	}

	docMap := schemaToFlatMap(doc)
	prevDocMap := schemaToFlatMap(prevDoc)

	db.indexes.Range(func(propName string, index *MemIndex) bool {
		index.Remove(id, prevDocMap[propName])
		return true
	})

	db.docs.Set(id, doc)

	db.indexes.Range(func(propName string, index *MemIndex) bool {
		index.Add(id, docMap[propName])
		return true
	})

	return Record[Schema]{Id: id, S: doc}, nil
}

func (db *MemDB[Schema]) Delete(id string) error {
	doc, ok := db.docs.Get(id)
	if !ok {
		return fmt.Errorf("document not found")
	}

	docMap := schemaToFlatMap(doc)

	db.indexes.Range(func(propName string, index *MemIndex) bool {
		index.Remove(id, docMap[propName])
		return true
	})

	db.docs.Del(id)

	return nil
}

func (db *MemDB[Schema]) Search(params SearchParams) []Record[Schema] {
	recordsIds := make(map[string]struct{})
	records := make([]Record[Schema], 0)
	props := params.Properties

	if len(params.Properties) == 1 && params.Properties[0] == WILDCARD {
		props = make([]string, 0)
		var s Schema
		for key := range schemaToFlatMap(s) {
			props = append(props, key)
		}
	}

	for _, prop := range props {
		if index, ok := db.indexes.Get(prop); ok {
			ids := index.Find(FindParams{
				Query:    params.Query,
				BoolMode: params.BoolMode,
			})
			for _, id := range ids {
				recordsIds[id] = struct{}{}
			}
		}
	}

	for id := range recordsIds {
		doc, _ := db.docs.Get(id)
		records = append(records, Record[Schema]{Id: id, S: doc})
	}

	return records
}

func NewIndex() *MemIndex {
	return &MemIndex{
		hashmap.New[string, *hashmap.Map[string, RecordInfo]](),
	}
}

func (idx *MemIndex) Add(id string, text string) {
	tokens := lib.Tokenize(text)
	tokensCount := lib.Count(tokens)

	for token, count := range tokensCount {
		recordsInfos, _ := idx.GetOrInsert(token, hashmap.New[string, RecordInfo]())
		recordsInfos.Insert(id, RecordInfo{count})
	}
}

func (idx *MemIndex) Remove(id string, text string) {
	tokens := lib.Tokenize(text)

	for _, token := range tokens {
		if recordsInfos, ok := idx.Get(token); ok {
			recordsInfos.Del(id)
		}
	}
}

func (idx *MemIndex) Find(params FindParams) []string {
	tokens := lib.Tokenize(params.Query)
	recordsIds := make(map[string]int)
	docIds := make([]string, 0)

	for _, token := range tokens {
		if infos, ok := idx.Get(token); ok {
			infos.Range(func(id string, info RecordInfo) bool {
				recordsIds[id] += 1
				return true
			})
		}
	}

	for id, tokensCount := range recordsIds {
		if (params.BoolMode == AND && tokensCount == len(tokens)) ||
			params.BoolMode == OR {
			docIds = append(docIds, id)
		}
	}

	return docIds
}

func schemaToFlatMap(obj any, prefix ...string) map[string]string {
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
				for key, value := range schemaToFlatMap(v.Field(i).Interface(), propName) {
					m[key] = value
				}
			} else {
				m[propName] = v.Field(i).String()
			}
		}
	}

	return m
}
