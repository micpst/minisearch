package store

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/micpst/minisearch/pkg/lib"
	"github.com/micpst/minisearch/pkg/tokenizer"
)

const (
	AND Mode = "AND"
	OR  Mode = "OR"
)

type Mode string

type SchemaProps any

type Record[Schema SchemaProps] struct {
	Id   string
	Data Schema
}

type InsertParams[Schema SchemaProps] struct {
	Document Schema
	Language tokenizer.Language
}

type InsertBatchParams[Schema SchemaProps] struct {
	Documents []Schema
	BatchSize int
	Language  tokenizer.Language
}

type UpdateParams[Schema SchemaProps] struct {
	Id       string
	Document Schema
	Language tokenizer.Language
}

type DeleteParams[Schema SchemaProps] struct {
	Id       string
	Language tokenizer.Language
}

type SearchParams struct {
	Query      string             `json:"query" binding:"required"`
	Properties []string           `json:"properties"`
	BoolMode   Mode               `json:"boolMode"`
	Exact      bool               `json:"exact"`
	Tolerance  int                `json:"tolerance"`
	Relevance  BM25Params         `json:"relevance"`
	Offset     int                `json:"offset"`
	Limit      int                `json:"limit"`
	Language   tokenizer.Language `json:"lang"`
}

type BM25Params struct {
	K float64 `json:"k"`
	B float64 `json:"b"`
	D float64 `json:"d"`
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

type Config struct {
	DefaultLanguage tokenizer.Language
	TokenizerConfig *tokenizer.Config
}

type MemDB[Schema SchemaProps] struct {
	mutex           sync.RWMutex
	documents       map[string]Schema
	indexes         map[string]*Index
	indexKeys       []string
	defaultLanguage tokenizer.Language
	tokenizerConfig *tokenizer.Config
}

func New[Schema SchemaProps](c *Config) *MemDB[Schema] {
	db := &MemDB[Schema]{
		documents:       make(map[string]Schema),
		indexes:         make(map[string]*Index),
		indexKeys:       make([]string, 0),
		defaultLanguage: c.DefaultLanguage,
		tokenizerConfig: c.TokenizerConfig,
	}
	db.buildIndexes()
	return db
}

func (db *MemDB[Schema]) buildIndexes() {
	var s Schema
	for key := range flattenSchema(s) {
		db.indexes[key] = NewIndex()
		db.indexKeys = append(db.indexKeys, key)
	}
}

func (db *MemDB[Schema]) Insert(params *InsertParams[Schema]) (Record[Schema], error) {
	id := uuid.NewString()
	document := flattenSchema(params.Document)

	language := params.Language
	if language == "" {
		language = db.defaultLanguage

	} else if !tokenizer.IsSupportedLanguage(language) {
		return Record[Schema]{}, fmt.Errorf("not supported language")
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, ok := db.documents[id]; ok {
		return Record[Schema]{}, fmt.Errorf("document id already exists")
	}

	db.documents[id] = params.Document
	db.indexDocument(id, document, language)

	return Record[Schema]{Id: id, Data: params.Document}, nil
}

func (db *MemDB[Schema]) InsertBatch(params *InsertBatchParams[Schema]) []error {
	batchCount := int(math.Ceil(float64(len(params.Documents)) / float64(params.BatchSize)))
	docsChan := make(chan Schema)
	errsChan := make(chan error)

	var wg sync.WaitGroup
	wg.Add(batchCount)

	go func() {
		for _, doc := range params.Documents {
			docsChan <- doc
		}
		close(docsChan)
	}()

	for i := 0; i < batchCount; i++ {
		go func() {
			defer wg.Done()
			for doc := range docsChan {
				insertParams := InsertParams[Schema]{
					Document: doc,
					Language: params.Language,
				}
				if _, err := db.Insert(&insertParams); err != nil {
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

func (db *MemDB[Schema]) Update(params *UpdateParams[Schema]) (Record[Schema], error) {
	document := flattenSchema(params.Document)

	language := params.Language
	if language == "" {
		language = db.defaultLanguage

	} else if !tokenizer.IsSupportedLanguage(language) {
		return Record[Schema]{}, fmt.Errorf("not supported language")
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	oldDocument, ok := db.documents[params.Id]
	if !ok {
		return Record[Schema]{}, fmt.Errorf("document not found")
	}

	db.indexDocument(params.Id, document, language)
	document = flattenSchema(oldDocument)
	db.deindexDocument(params.Id, document, language)

	db.documents[params.Id] = params.Document

	return Record[Schema]{Id: params.Id, Data: params.Document}, nil
}

func (db *MemDB[Schema]) Delete(params *DeleteParams[Schema]) error {
	language := params.Language
	if language == "" {
		language = db.defaultLanguage

	} else if !tokenizer.IsSupportedLanguage(language) {
		return fmt.Errorf("not supported language")
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	document, ok := db.documents[params.Id]
	if !ok {
		return fmt.Errorf("document not found")
	}

	doc := flattenSchema(document)
	db.deindexDocument(params.Id, doc, language)

	delete(db.documents, params.Id)

	return nil
}

func (db *MemDB[Schema]) Search(params *SearchParams) (SearchResult[Schema], error) {
	allIdScores := make(map[string]float64)
	results := make(SearchHits[Schema], 0)

	properties := params.Properties
	if len(params.Properties) == 0 {
		properties = db.indexKeys
	}

	language := params.Language
	if params.Language == "" {
		language = db.defaultLanguage

	} else if !tokenizer.IsSupportedLanguage(language) {
		return SearchResult[Schema]{}, fmt.Errorf("not supported language")
	}

	tokens, _ := tokenizer.Tokenize(&tokenizer.TokenizeParams{
		Text:            params.Query,
		Language:        language,
		AllowDuplicates: false,
	}, db.tokenizerConfig)

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	for _, prop := range properties {
		if index, ok := db.indexes[prop]; ok {
			idScores := index.Find(&FindParams{
				Tokens:    tokens,
				BoolMode:  params.BoolMode,
				Exact:     params.Exact,
				Tolerance: params.Tolerance,
				Relevance: params.Relevance,
				DocsCount: len(db.documents),
			})
			for id, score := range idScores {
				allIdScores[id] += score
			}
		}
	}

	for id, score := range allIdScores {
		if doc, ok := db.documents[id]; ok {
			results = append(results, SearchHit[Schema]{
				Id:    id,
				Data:  doc,
				Score: score,
			})
		}
	}

	sort.Sort(results)

	start, stop := lib.Paginate(params.Offset, params.Limit, len(results))

	return SearchResult[Schema]{Hits: results[start:stop], Count: len(results)}, nil
}

func (db *MemDB[Schema]) indexDocument(id string, document map[string]string, language tokenizer.Language) {
	for propName, index := range db.indexes {
		tokens, _ := tokenizer.Tokenize(&tokenizer.TokenizeParams{
			Text:            document[propName],
			Language:        language,
			AllowDuplicates: true,
		}, db.tokenizerConfig)

		index.Insert(&IndexParams{
			Id:        id,
			Tokens:    tokens,
			DocsCount: len(db.documents),
		})
	}
}

func (db *MemDB[Schema]) deindexDocument(id string, document map[string]string, language tokenizer.Language) {
	for propName, index := range db.indexes {
		tokens, _ := tokenizer.Tokenize(&tokenizer.TokenizeParams{
			Text:            document[propName],
			Language:        language,
			AllowDuplicates: false,
		}, db.tokenizerConfig)

		index.Delete(&IndexParams{
			Id:        id,
			Tokens:    tokens,
			DocsCount: len(db.documents),
		})
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
