package storage

import (
	"fmt"
	"reflect"

	"github.com/cornelk/hashmap"
	"github.com/google/uuid"
	"github.com/micpst/full-text-search-engine/src/lib"
)

type DocInfo struct {
	docId string
	freq  uint32
}

type MemDB[Schema any] struct {
	docs  *hashmap.Map[string, Schema]
	index *hashmap.Map[string, []DocInfo]
}

func New[Schema any]() *MemDB[Schema] {
	return &MemDB[Schema]{
		docs:  hashmap.New[string, Schema](),
		index: hashmap.New[string, []DocInfo](),
	}
}

func (db *MemDB[Schema]) Create(doc Schema) (string, error) {
	id := uuid.NewString()
	if ok := db.docs.Insert(id, doc); !ok {
		return "", fmt.Errorf("document cannot be created")
	}

	fields := getIndexFields(doc)
	for _, field := range fields {
		db.indexField(id, field)
	}

	return id, nil
}

func (db *MemDB[Schema]) Update(id string, doc Schema) error {
	prevDoc, ok := db.docs.Get(id)
	if !ok {
		return fmt.Errorf("document not found")
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

	return nil
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

func (db *MemDB[Schema]) Search(query string) []Schema {
	docs := []Schema{}
	infos := []DocInfo{}
	tokens := lib.Tokenize(query)

	for _, token := range tokens {
		docsInfo, _ := db.index.Get(token)

		for _, info := range docsInfo {
			if idx := getDocumentInfoIndex(infos, info.docId); idx >= 0 {
				infos[idx].freq += info.freq
			} else {
				infos = append(infos, info)
			}
		}
	}

	for _, info := range infos {
		doc, _ := db.docs.Get(info.docId)
		docs = append(docs, doc)
	}

	return docs
}

func (db *MemDB[Schema]) indexField(id string, text string) {
	tokens := lib.Tokenize(text)
	tokensCount := lib.Count(tokens)

	for token, count := range tokensCount {
		docsInfo, _ := db.index.GetOrInsert(token, []DocInfo{})
		docsInfo = append(docsInfo, DocInfo{id, count})
		db.index.Set(token, docsInfo)
	}
}

func (db *MemDB[Schema]) deindexField(id string, text string) {
	tokens := lib.Tokenize(text)

	for _, token := range tokens {
		if docsInfo, ok := db.index.Get(token); ok {
			var newDocsInfo []DocInfo
			for _, info := range docsInfo {
				if info.docId != id {
					newDocsInfo = append(newDocsInfo, info)
				}
			}
			db.index.Set(token, newDocsInfo)
		}
	}
}

func getIndexFields(d any) []string {
	fields := make([]string, 0)
	val := reflect.ValueOf(d)
	t := reflect.TypeOf(d)

	for i := 0; i < val.NumField(); i++ {
		f := t.Field(i)
		if v, ok := f.Tag.Lookup("index"); ok && v == "true" {
			fields = append(fields, val.Field(i).String())
		}
	}

	return fields
}

func getDocumentInfoIndex(docsInfos []DocInfo, docId string) int {
	for idx, info := range docsInfos {
		if info.docId == docId {
			return idx
		}
	}
	return -1
}
