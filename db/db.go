package db

import (
	"fmt"

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

func AddDocument(content string) (*Document, error) {
	id := uuid.Must(uuid.NewV4()).String()
	newDoc := Document{id, content}

	if ok := documents.Insert(newDoc.Id, newDoc.Content); !ok {
		return nil, fmt.Errorf("document cannot be created")
	}
	indexDocument(&newDoc)

	return &newDoc, nil
}

func ModifyDocument(id string, newContent string) (*Document, error) {
	newDoc := Document{id, newContent}
	content, ok := documents.Get(id)
	if !ok {
		return nil, fmt.Errorf("document not found")
	}

	deindexDocument(&Document{id, content})
	documents.Set(newDoc.Id, newDoc.Content)
	indexDocument(&newDoc)

	return &newDoc, nil
}

func RemoveDocument(id string) error {
	content, ok := documents.Get(id)
	if !ok {
		return fmt.Errorf("document not found")
	}

	deindexDocument(&Document{id, content})
	documents.Del(id)

	return nil
}

func indexDocument(d *Document) {
	tokens := parser.Tokenize(d.Content)
	tokensCount := parser.Count(tokens)

	for token, count := range tokensCount {
		docsInfo, _ := index.GetOrInsert(token, []DocumentInfo{})
		docsInfo = append(docsInfo, DocumentInfo{d.Id, count})
		index.Set(token, docsInfo)
	}
}

func deindexDocument(d *Document) {
	tokens := parser.Tokenize(d.Content)

	for _, token := range tokens {
		if docsInfo, ok := index.Get(token); ok {
			var newDocsInfo []DocumentInfo
			for _, info := range docsInfo {
				if info.DocumentId != d.Id {
					newDocsInfo = append(newDocsInfo, info)
				}
			}
			index.Set(token, newDocsInfo)
		}
	}
}
