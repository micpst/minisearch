package store

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/cornelk/hashmap"
	"github.com/google/uuid"
	"github.com/micpst/full-text-search-engine/src/lib"
)

type SchemaProps any

type Record[Schema SchemaProps] struct {
	Id string
	S  Schema
}

type RecordInfo struct {
	freq uint32
}

type SearchParams struct {
	Query string
	Exact bool
}

type MemDB[Schema SchemaProps] struct {
	docs  *hashmap.Map[string, Schema]
	index *hashmap.Map[string, *hashmap.Map[string, RecordInfo]]
}

func New[Schema SchemaProps]() *MemDB[Schema] {
	return &MemDB[Schema]{
		docs:  hashmap.New[string, Schema](),
		index: hashmap.New[string, *hashmap.Map[string, RecordInfo]](),
	}
}

func (db *MemDB[Schema]) Insert(doc Schema) (Record[Schema], error) {
	id := uuid.NewString()
	if ok := db.docs.Insert(id, doc); !ok {
		return Record[Schema]{}, fmt.Errorf("document cannot be created")
	}

	db.indexDocument(id, doc)

	return Record[Schema]{id, doc}, nil
}

func (db *MemDB[Schema]) InsertBatch(docs []Schema, batchSize int) []error {
	n := len(docs) / batchSize
	in := make(chan Schema)
	out := make(chan error)

	var wg sync.WaitGroup
	wg.Add(n)

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
		for _, d := range docs {
			in <- d
		}
		close(in)
	}()
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

	db.deindexDocument(id, prevDoc)
	db.docs.Set(id, doc)
	db.indexDocument(id, doc)

	return Record[Schema]{id, doc}, nil
}

func (db *MemDB[Schema]) Delete(id string) error {
	doc, ok := db.docs.Get(id)
	if !ok {
		return fmt.Errorf("document not found")
	}

	db.deindexDocument(id, doc)
	db.docs.Del(id)

	return nil
}

func (db *MemDB[Schema]) Search(params SearchParams) []Record[Schema] {
	recordsIds := make(map[string]int)
	records := make([]Record[Schema], 0)
	tokens := lib.Tokenize(params.Query)

	for _, token := range tokens {
		infos, _ := db.index.Get(token)
		infos.Range(func(id string, info RecordInfo) bool {
			recordsIds[id] += 1
			return true
		})
	}

	for id, tokensCount := range recordsIds {
		if !params.Exact || tokensCount == len(tokens) {
			doc, _ := db.docs.Get(id)
			records = append(records, Record[Schema]{id, doc})
		}
	}

	return records
}

func (db *MemDB[Schema]) indexDocument(id string, doc Schema) {
	text := strings.Join(getIndexFields(doc), " ")
	tokens := lib.Tokenize(text)
	tokensCount := lib.Count(tokens)

	for token, count := range tokensCount {
		recordsInfos, _ := db.index.GetOrInsert(token, hashmap.New[string, RecordInfo]())
		recordsInfos.Insert(id, RecordInfo{count})
	}
}

func (db *MemDB[Schema]) deindexDocument(id string, doc Schema) {
	text := strings.Join(getIndexFields(doc), " ")
	tokens := lib.Tokenize(text)

	for _, token := range tokens {
		if recordsInfos, ok := db.index.Get(token); ok {
			recordsInfos.Del(id)
		}
	}
}

func getIndexFields(obj any) []string {
	fields := make([]string, 0)
	val := reflect.ValueOf(obj)
	t := reflect.TypeOf(obj)

	for i := 0; i < val.NumField(); i++ {
		f := t.Field(i)
		if v, ok := f.Tag.Lookup("index"); ok && v == "true" {
			fields = append(fields, val.Field(i).String())
		}
	}

	return fields
}
