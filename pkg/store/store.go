package store

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/micpst/minisearch/pkg/invindex"
	"github.com/micpst/minisearch/pkg/lib"
	"github.com/micpst/minisearch/pkg/tokenizer"
)

const (
	AND Mode = "AND"
	OR  Mode = "OR"
)

const WILDCARD = "*"

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

type SearchParams struct {
	Query      string
	Properties []string
	BoolMode   Mode
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
	indexes         map[string]invindex.InvIndex
	indexKeys       []string
	occurrences     map[string]map[string]int
	defaultLanguage tokenizer.Language
	tokenizer       *tokenizer.Tokenizer
}

func New[Schema SchemaProps](c *Config) *MemDB[Schema] {
	db := &MemDB[Schema]{
		documents:       make(map[string]Schema),
		indexes:         make(map[string]invindex.InvIndex),
		indexKeys:       make([]string, 0),
		occurrences:     make(map[string]map[string]int),
		defaultLanguage: c.DefaultLanguage,
		tokenizer:       tokenizer.New(c.TokenizerConfig),
	}
	db.buildIndexes()
	return db
}

func (db *MemDB[Schema]) buildIndexes() {
	var s Schema
	for key := range flattenSchema(s) {
		db.indexes[key] = invindex.InvIndex{}
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

	} else if !db.tokenizer.IsSupportedLanguage(idxParams.language) {
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

	} else if !db.tokenizer.IsSupportedLanguage(params.Language) {
		return Record[Schema]{}, fmt.Errorf("not supported language")
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	oldDocument, ok := db.documents[params.Id]
	if !ok {
		return Record[Schema]{}, fmt.Errorf("document not found")
	}

	db.indexDocument(&idxParams)
	idxParams.document = flattenSchema(oldDocument)
	db.deindexDocument(&idxParams)

	db.documents[params.Id] = params.Document

	return Record[Schema]{Id: params.Id, Data: params.Document}, nil
}

func (db *MemDB[Schema]) Delete(params *DeleteParams[Schema]) error {
	idxParams := indexParams{
		id:       params.Id,
		language: params.Language,
	}

	if idxParams.language == "" {
		idxParams.language = db.defaultLanguage

	} else if !db.tokenizer.IsSupportedLanguage(params.Language) {
		return fmt.Errorf("not supported language")
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	document, ok := db.documents[params.Id]
	if !ok {
		return fmt.Errorf("document not found")
	}

	idxParams.document = flattenSchema(document)
	db.deindexDocument(&idxParams)

	delete(db.documents, params.Id)

	return nil
}

func (db *MemDB[Schema]) Search(params *SearchParams) (SearchResult[Schema], error) {
	idScores := make(map[string]float64)
	results := make(SearchHits[Schema], 0)

	props := params.Properties
	if len(props) == 1 && props[0] == WILDCARD {
		props = db.indexKeys
	}

	language := params.Language
	if language == "" {
		language = db.defaultLanguage
	}

	input := tokenizer.TokenizeInput{
		Text:            params.Query,
		Language:        language,
		AllowDuplicates: false,
	}
	tokens, err := db.tokenizer.Tokenize(&input)
	if err != nil {
		return SearchResult[Schema]{Hits: results}, err
	}

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	for _, prop := range props {
		if index, ok := db.indexes[prop]; ok {
			idTokensCount := make(map[string]int)

			for _, token := range tokens {
				idInfos := index.Find(token)
				for id, info := range idInfos {
					idScores[id] += lib.TfIdf(info.TermFrequency, db.occurrences[prop][token], len(db.documents))
					idTokensCount[id]++
				}
			}

			for id, tokensCount := range idTokensCount {
				if params.BoolMode == AND && tokensCount != len(tokens) {
					delete(idScores, id)
				}
			}
		}
	}

	for id, score := range idScores {
		if doc, ok := db.documents[id]; ok {
			results = append(results, SearchHit[Schema]{Id: id, Data: doc, Score: score})
		}
	}

	sort.Sort(results)

	start, stop := lib.Paginate(params.Offset, params.Limit, len(results))

	return SearchResult[Schema]{
		Hits:  results[start:stop],
		Count: len(results),
	}, nil
}

func (db *MemDB[Schema]) indexDocument(params *indexParams) {
	input := tokenizer.TokenizeInput{
		Language:        params.language,
		AllowDuplicates: true,
	}

	for propName, index := range db.indexes {
		input.Text = params.document[propName]
		tokens, _ := db.tokenizer.Tokenize(&input)
		tokensCount := lib.Count(tokens)

		for token, count := range tokensCount {
			tokenFrequency := float64(count) / float64(len(tokens))
			index.Add(params.id, token, tokenFrequency)

			db.occurrences[propName][token]++
		}
	}
}

func (db *MemDB[Schema]) deindexDocument(params *indexParams) {
	input := tokenizer.TokenizeInput{
		Language:        params.language,
		AllowDuplicates: false,
	}

	for propName, index := range db.indexes {
		input.Text = params.document[propName]
		tokens, _ := db.tokenizer.Tokenize(&input)

		for _, token := range tokens {
			index.Remove(params.id, token)

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
