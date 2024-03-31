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

type Schema any

type Record[S Schema] struct {
	Id   string
	Data S
}

type InsertParams[S Schema] struct {
	Document S
	Language tokenizer.Language
}

type InsertBatchParams[S Schema] struct {
	Documents []S
	BatchSize int
	Language  tokenizer.Language
}

type UpdateParams[S Schema] struct {
	Id       string
	Document S
	Language tokenizer.Language
}

type DeleteParams[S Schema] struct {
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

type SearchResult[S Schema] struct {
	Hits  SearchHits[S]
	Count int
}

type SearchHit[S Schema] struct {
	Id    string
	Data  S
	Score float64
}

type SearchHits[S Schema] []SearchHit[S]

func (r SearchHits[S]) Len() int { return len(r) }

func (r SearchHits[S]) Swap(i, j int) { r[i], r[j] = r[j], r[i] }

func (r SearchHits[S]) Less(i, j int) bool { return r[i].Score > r[j].Score }

type Config struct {
	DefaultLanguage tokenizer.Language
	TokenizerConfig *tokenizer.Config
}

type MemDB[S Schema] struct {
	mutex           sync.RWMutex
	documents       map[string]S
	indexes         map[string]*Index
	indexKeys       []string
	defaultLanguage tokenizer.Language
	tokenizerConfig *tokenizer.Config
}

func New[S Schema](c *Config) *MemDB[S] {
	db := &MemDB[S]{
		documents:       make(map[string]S),
		indexes:         make(map[string]*Index),
		indexKeys:       make([]string, 0),
		defaultLanguage: c.DefaultLanguage,
		tokenizerConfig: c.TokenizerConfig,
	}
	db.buildIndexes()
	return db
}

func (db *MemDB[S]) buildIndexes() {
	var s S
	for key := range flattenSchema(s) {
		db.indexes[key] = NewIndex()
		db.indexKeys = append(db.indexKeys, key)
	}
}

func (db *MemDB[S]) Insert(params *InsertParams[S]) (Record[S], error) {
	id := uuid.NewString()
	document := flattenSchema(params.Document)

	language := params.Language
	if language == "" {
		language = db.defaultLanguage

	} else if !tokenizer.IsSupportedLanguage(language) {
		return Record[S]{}, &tokenizer.LanguageNotSupportedError{Language: language}
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, ok := db.documents[id]; ok {
		return Record[S]{}, &DocumentAlreadyExistsError{Id: id}
	}

	db.documents[id] = params.Document
	db.indexDocument(id, document, language)

	return Record[S]{Id: id, Data: params.Document}, nil
}

func (db *MemDB[S]) InsertBatch(params *InsertBatchParams[S]) []error {
	batchCount := int(math.Ceil(float64(len(params.Documents)) / float64(params.BatchSize)))
	docsChan := make(chan S)
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
				insertParams := InsertParams[S]{
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

func (db *MemDB[S]) Update(params *UpdateParams[S]) (Record[S], error) {
	document := flattenSchema(params.Document)

	language := params.Language
	if language == "" {
		language = db.defaultLanguage

	} else if !tokenizer.IsSupportedLanguage(language) {
		return Record[S]{}, &tokenizer.LanguageNotSupportedError{Language: language}
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	oldDocument, ok := db.documents[params.Id]
	if !ok {
		return Record[S]{}, &DocumentNotFoundError{Id: params.Id}
	}

	db.indexDocument(params.Id, document, language)
	document = flattenSchema(oldDocument)
	db.deindexDocument(params.Id, document, language)

	db.documents[params.Id] = params.Document

	return Record[S]{Id: params.Id, Data: params.Document}, nil
}

func (db *MemDB[S]) Delete(params *DeleteParams[S]) error {
	language := params.Language
	if language == "" {
		language = db.defaultLanguage

	} else if !tokenizer.IsSupportedLanguage(language) {
		return &tokenizer.LanguageNotSupportedError{Language: language}
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	document, ok := db.documents[params.Id]
	if !ok {
		return &DocumentNotFoundError{Id: params.Id}
	}

	doc := flattenSchema(document)
	db.deindexDocument(params.Id, doc, language)

	delete(db.documents, params.Id)

	return nil
}

func (db *MemDB[S]) Search(params *SearchParams) (SearchResult[S], error) {
	allIdScores := make(map[string]float64)
	results := make(SearchHits[S], 0)

	properties := params.Properties
	if len(params.Properties) == 0 {
		properties = db.indexKeys
	}

	language := params.Language
	if params.Language == "" {
		language = db.defaultLanguage

	} else if !tokenizer.IsSupportedLanguage(language) {
		return SearchResult[S]{}, &tokenizer.LanguageNotSupportedError{Language: language}
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
			results = append(results, SearchHit[S]{
				Id:    id,
				Data:  doc,
				Score: score,
			})
		}
	}

	sort.Sort(results)

	start, stop := lib.Paginate(params.Offset, params.Limit, len(results))

	return SearchResult[S]{Hits: results[start:stop], Count: len(results)}, nil
}

func (db *MemDB[S]) indexDocument(id string, document map[string]string, language tokenizer.Language) {
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

func (db *MemDB[S]) deindexDocument(id string, document map[string]string, language tokenizer.Language) {
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
