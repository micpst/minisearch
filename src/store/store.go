package store

import (
	"fmt"
	"reflect"
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
	recId string
	freq  uint32
}

type MemDB[Schema SchemaProps] struct {
	docs  *hashmap.Map[string, Schema]
	index *hashmap.Map[string, []RecordInfo]
}

func New[Schema SchemaProps]() *MemDB[Schema] {
	return &MemDB[Schema]{
		docs:  hashmap.New[string, Schema](),
		index: hashmap.New[string, []RecordInfo](),
	}
}

func (db *MemDB[Schema]) Insert(doc Schema) (Record[Schema], error) {
	id := uuid.NewString()
	if ok := db.docs.Insert(id, doc); !ok {
		return Record[Schema]{}, fmt.Errorf("document cannot be created")
	}

	fields := getIndexFields(doc)
	for _, field := range fields {
		db.indexField(id, field)
	}

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

	db.docs.Set(id, doc)

	fields := getIndexFields(prevDoc)
	for _, field := range fields {
		db.deindexField(id, field)
	}

	fields = getIndexFields(doc)
	for _, field := range fields {
		db.indexField(id, field)
	}

	return Record[Schema]{id, doc}, nil
}

func (db *MemDB[Schema]) Delete(id string) error {
	doc, ok := db.docs.Get(id)
	if !ok {
		return fmt.Errorf("document not found")
	}

	db.docs.Del(id)

	fields := getIndexFields(doc)
	for _, field := range fields {
		db.deindexField(id, field)
	}

	return nil
}

func (db *MemDB[Schema]) Search(query string) []Record[Schema] {
	records := make([]Record[Schema], 0)
	infos := make([]RecordInfo, 0)
	tokens := lib.Tokenize(query)

	for _, token := range tokens {
		recordsInfos, _ := db.index.Get(token)

		for _, info := range recordsInfos {
			if idx := findRecordInfo(infos, info.recId); idx >= 0 {
				infos[idx].freq += info.freq
			} else {
				infos = append(infos, info)
			}
		}
	}

	for _, info := range infos {
		doc, _ := db.docs.Get(info.recId)
		records = append(records, Record[Schema]{info.recId, doc})
	}

	return records
}

func (db *MemDB[Schema]) indexField(id string, text string) {
	tokens := lib.Tokenize(text)
	tokensCount := lib.Count(tokens)

	for token, count := range tokensCount {
		recordsInfos, _ := db.index.GetOrInsert(token, []RecordInfo{})
		recordsInfos = append(recordsInfos, RecordInfo{id, count})
		db.index.Set(token, recordsInfos)
	}
}

func (db *MemDB[Schema]) deindexField(id string, text string) {
	tokens := lib.Tokenize(text)

	for _, token := range tokens {
		if recordsInfos, ok := db.index.Get(token); ok {
			var newRecordsInfos []RecordInfo
			for _, info := range recordsInfos {
				if info.recId != id {
					newRecordsInfos = append(newRecordsInfos, info)
				}
			}
			db.index.Set(token, newRecordsInfos)
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

func findRecordInfo(infos []RecordInfo, id string) int {
	for idx, info := range infos {
		if info.recId == id {
			return idx
		}
	}
	return -1
}
