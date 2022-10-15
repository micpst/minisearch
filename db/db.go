package db

import (
	"github.com/cornelk/hashmap"
	"github.com/gofrs/uuid"
	"github.com/micpst/full-text-search-engine/parser"
)

type Document struct {
	Id      string
	Content string
}

type DocumentInfo struct {
	DocumentId string
	Frequency  uint32
}

var (
	documents *hashmap.Map[string, string]
	index     *hashmap.Map[string, []DocumentInfo]
)

func init() {
	documents = hashmap.New[string, string]()
	index = hashmap.New[string, []DocumentInfo]()
}

func AddDocument(data string) *Document {
	id := uuid.Must(uuid.NewV4()).String()
	documents.Insert(id, data)
	return &Document{id, data}
}

func UpdateDocument(d *Document) error {
	return nil
}

func DeleteDocument(id string) error {
	return nil
}

func IndexDocument(d *Document) {
	tokens := parser.Tokenize(d.Content)
	tokensCount := parser.Count(tokens)

	for token, count := range tokensCount {
		docsInfo, _ := index.GetOrInsert(token, []DocumentInfo{})
		docsInfo = append(docsInfo, DocumentInfo{d.Id, count})
		index.Set(token, docsInfo)
	}
}
