package store

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/micpst/minisearch/pkg/lib"
	"github.com/micpst/minisearch/pkg/radix"
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

type indexParams struct {
	id       string
	document map[string]string
	language tokenizer.Language
}

type findParams struct {
	query      string
	properties []string
	boolMode   Mode
	exact      bool
	tolerance  int
	language   tokenizer.Language
}

type SearchParams struct {
	Query      string
	Properties []string
	BoolMode   Mode
	Exact      bool
	Tolerance  int
	Offset     int
	Limit      int
	Language   tokenizer.Language
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
	indexes         map[string]*radix.Trie
	indexKeys       []string
	occurrences     map[string]map[string]int
	defaultLanguage tokenizer.Language
	tokenizerConfig *tokenizer.Config
}

func New[Schema SchemaProps](c *Config) *MemDB[Schema] {
	db := &MemDB[Schema]{
		documents:       make(map[string]Schema),
		indexes:         make(map[string]*radix.Trie),
		indexKeys:       make([]string, 0),
		occurrences:     make(map[string]map[string]int),
		defaultLanguage: c.DefaultLanguage,
		tokenizerConfig: c.TokenizerConfig,
	}
	db.buildIndexes()
	return db
}

func (db *MemDB[Schema]) buildIndexes() {
	var s Schema
	for key := range flattenSchema(s) {
		db.indexes[key] = radix.New()
		db.indexKeys = append(db.indexKeys, key)
		db.occurrences[key] = make(map[string]int)
	}
}

func (db *MemDB[Schema]) Insert(params *InsertParams[Schema]) (Record[Schema], error) {
	idxParams := indexParams{
		id:       uuid.NewString(),
		document: flattenSchema(params.Document),
		language: params.Language,
	}

	if idxParams.language == "" {
		idxParams.language = db.defaultLanguage

	} else if !tokenizer.IsSupportedLanguage(idxParams.language) {
		return Record[Schema]{}, fmt.Errorf("not supported language")
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, ok := db.documents[idxParams.id]; ok {
		return Record[Schema]{}, fmt.Errorf("document id already exists")
	}

	db.documents[idxParams.id] = params.Document
	db.indexDocument(&idxParams)

	return Record[Schema]{Id: idxParams.id, Data: params.Document}, nil
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
	idxParams := indexParams{
		id:       params.Id,
		language: params.Language,
		document: flattenSchema(params.Document),
	}

	if idxParams.language == "" {
		idxParams.language = db.defaultLanguage

	} else if !tokenizer.IsSupportedLanguage(idxParams.language) {
		return Record[Schema]{}, fmt.Errorf("not supported language")
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	oldDocument, ok := db.documents[idxParams.id]
	if !ok {
		return Record[Schema]{}, fmt.Errorf("document not found")
	}

	db.indexDocument(&idxParams)
	idxParams.document = flattenSchema(oldDocument)
	db.deindexDocument(&idxParams)

	db.documents[idxParams.id] = params.Document

	return Record[Schema]{Id: idxParams.id, Data: params.Document}, nil
}

func (db *MemDB[Schema]) Delete(params *DeleteParams[Schema]) error {
	idxParams := indexParams{
		id:       params.Id,
		language: params.Language,
	}

	if idxParams.language == "" {
		idxParams.language = db.defaultLanguage

	} else if !tokenizer.IsSupportedLanguage(idxParams.language) {
		return fmt.Errorf("not supported language")
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	document, ok := db.documents[idxParams.id]
	if !ok {
		return fmt.Errorf("document not found")
	}

	idxParams.document = flattenSchema(document)
	db.deindexDocument(&idxParams)

	delete(db.documents, idxParams.id)

	return nil
}

func (db *MemDB[Schema]) Search(params *SearchParams) (SearchResult[Schema], error) {
	idxParams := findParams{
		query:      params.Query,
		properties: params.Properties,
		boolMode:   params.BoolMode,
		exact:      params.Exact,
		tolerance:  params.Tolerance,
		language:   params.Language,
	}

	if len(idxParams.properties) == 0 {
		idxParams.properties = db.indexKeys
	}

	if idxParams.language == "" {
		idxParams.language = db.defaultLanguage

	} else if !tokenizer.IsSupportedLanguage(idxParams.language) {
		return SearchResult[Schema]{}, fmt.Errorf("not supported language")
	}

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	results := make(SearchHits[Schema], 0)
	idScores := db.findDocumentIds(&idxParams)

	for id, score := range idScores {
		if doc, ok := db.documents[id]; ok {
			results = append(results, SearchHit[Schema]{Id: id, Data: doc, Score: score})
		}
	}

	sort.Sort(results)

	start, stop := lib.Paginate(params.Offset, params.Limit, len(results))

	return SearchResult[Schema]{Hits: results[start:stop], Count: len(results)}, nil
}

func (db *MemDB[Schema]) findDocumentIds(params *findParams) map[string]float64 {
	tokenParams := tokenizer.TokenizeParams{
		Text:            params.query,
		Language:        params.language,
		AllowDuplicates: false,
	}
	tokens, _ := tokenizer.Tokenize(&tokenParams, db.tokenizerConfig)

	idScores := make(map[string]float64)
	for _, prop := range params.properties {
		if index, ok := db.indexes[prop]; ok {
			idTokensCount := make(map[string]int)

			for _, token := range tokens {
				infos := index.Find(&radix.FindParams{
					Term:      token,
					Tolerance: params.tolerance,
					Exact:     params.exact,
				})
				for _, info := range infos {
					idScores[info.Id] += lib.TfIdf(info.TermFrequency, db.occurrences[prop][token], len(db.documents))
					idTokensCount[info.Id]++
				}
			}

			for id, tokensCount := range idTokensCount {
				if params.boolMode == AND && tokensCount != len(tokens) {
					delete(idScores, id)
				}
			}
		}
	}

	return idScores
}

func (db *MemDB[Schema]) indexDocument(params *indexParams) {
	tokenParams := tokenizer.TokenizeParams{
		Language:        params.language,
		AllowDuplicates: true,
	}

	for propName, index := range db.indexes {
		tokenParams.Text = params.document[propName]
		tokens, _ := tokenizer.Tokenize(&tokenParams, db.tokenizerConfig)
		tokensCount := lib.Count(tokens)

		for token, count := range tokensCount {
			tokenFrequency := float64(count) / float64(len(tokens))
			index.Insert(&radix.InsertParams{
				Id:            params.id,
				Word:          token,
				TermFrequency: tokenFrequency,
			})

			db.occurrences[propName][token]++
		}
	}
}

func (db *MemDB[Schema]) deindexDocument(params *indexParams) {
	tokenParams := tokenizer.TokenizeParams{
		Language:        params.language,
		AllowDuplicates: false,
	}

	for propName, index := range db.indexes {
		tokenParams.Text = params.document[propName]
		tokens, _ := tokenizer.Tokenize(&tokenParams, db.tokenizerConfig)

		for _, token := range tokens {
			index.Delete(&radix.DeleteParams{
				Id:   params.id,
				Word: token,
			})

			db.occurrences[propName][token]--
			if db.occurrences[propName][token] == 0 {
				delete(db.occurrences[propName], token)
			}
		}
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
