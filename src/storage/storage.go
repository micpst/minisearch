package storage

import (
	"fmt"

	"github.com/cornelk/hashmap"
	"github.com/google/uuid"
	"github.com/micpst/full-text-search-engine/src/lib"
)

type Document struct {
	Id       string `json:"id"`
	Title    string `json:"title" xml:"title" binding:"required"`
	Url      string `json:"url" xml:"url" binding:"required"`
	Abstract string `json:"abstract" xml:"abstract" binding:"required"`
}

type DocumentInfo struct {
	docId string
	freq  uint32
}

type MemDB struct {
	docs  *hashmap.Map[string, Document]
	index *hashmap.Map[string, []DocumentInfo]
}

func (d *Document) Content() string {
	return fmt.Sprintf("%s %s", d.Title, d.Abstract)
}

func New() *MemDB {
	return &MemDB{
		docs:  hashmap.New[string, Document](),
		index: hashmap.New[string, []DocumentInfo](),
	}
}

func (db *MemDB) Create(doc Document) (*Document, error) {
	doc.Id = uuid.NewString()
	if ok := db.docs.Insert(doc.Id, doc); !ok {
		return nil, fmt.Errorf("document cannot be created")
	}

	db.indexDocument(&doc)

	return &doc, nil
}

func (db *MemDB) Update(id string, doc Document) (*Document, error) {
	prevDoc, ok := db.docs.Get(id)
	if !ok {
		return nil, fmt.Errorf("document not found")
	}

	db.deindexDocument(&prevDoc)
	db.docs.Set(id, doc)
	db.indexDocument(&doc)

	return &doc, nil
}

func (db *MemDB) Delete(id string) error {
	doc, ok := db.docs.Get(id)
	if !ok {
		return fmt.Errorf("document not found")
	}

	db.deindexDocument(&doc)
	db.docs.Del(id)

	return nil
}

func (db *MemDB) Search(query string) []Document {
	resultInfos := []DocumentInfo{}
	resultDocs := []Document{}
	tokens := lib.Tokenize(query)

	for _, token := range tokens {
		docsInfo, _ := db.index.Get(token)
		for _, info := range docsInfo {
			if idx := getDocumentInfoIndex(resultInfos, info.docId); idx >= 0 {
				resultInfos[idx].freq += info.freq
			} else {
				resultInfos = append(resultInfos, info)
			}
		}
	}

	for _, info := range resultInfos {
		doc, _ := db.docs.Get(info.docId)
		resultDocs = append(resultDocs, doc)
	}

	return resultDocs
}

func (db *MemDB) indexDocument(d *Document) {
	tokens := lib.Tokenize(d.Content())
	tokensCount := lib.Count(tokens)

	for token, count := range tokensCount {
		docsInfo, _ := db.index.GetOrInsert(token, []DocumentInfo{})
		docsInfo = append(docsInfo, DocumentInfo{d.Id, count})
		db.index.Set(token, docsInfo)
	}
}

func (db *MemDB) deindexDocument(d *Document) {
	tokens := lib.Tokenize(d.Content())

	for _, token := range tokens {
		if docsInfo, ok := db.index.Get(token); ok {
			var newDocsInfo []DocumentInfo
			for _, info := range docsInfo {
				if info.docId != d.Id {
					newDocsInfo = append(newDocsInfo, info)
				}
			}
			db.index.Set(token, newDocsInfo)
		}
	}
}

func getDocumentInfoIndex(docsInfos []DocumentInfo, docId string) int {
	for idx, info := range docsInfos {
		if info.docId == docId {
			return idx
		}
	}
	return -1
}
