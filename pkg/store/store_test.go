package store

import (
	"fmt"
	"log"
	"testing"

	"github.com/micpst/minisearch/pkg/tokenizer"
	"github.com/stretchr/testify/assert"
)

type TestCase[Given any, Expected any] struct {
	given    Given
	expected Expected
}

type IndexState struct {
	length      int
	occurrences int
}

type User struct {
	Name   string `index:"name"`
	Email  string `index:"email"`
	Joined string
}

type Document struct {
	Title    string `index:"title"`
	Abstract string `index:"abstract"`
	Url      string
	Author   User `index:"author"`
}

var testData = []User{
	{"Tom Haris", "tom@email.com", "2023-02-10T15:04:05Z07:00"},
	{"Jane", "myne123@email.com", "2023-02-10T15:04:06Z07:00"},
	{"Bob Brown", "bob03@email.com", "2023-02-10T15:04:07Z07:00"},
	{"Charlie Davis", "char@email.com", "2023-02-10T15:04:08Z07:00"},
	{"Tom Brown", "brown@email.com", "2023-02-10T15:04:09Z07:00"},
	{"Charlie Anderson", "anderson@email.com", "2023-02-10T15:04:10Z07:00"},
	{"Julia Hernandez", "juliah@email.com", "2023-02-10T15:04:11Z07:00"},
	{"Lucy Johnson", "lucy@email.com", "2023-02-10T15:04:12Z07:00"},
	{"Frank", "fischer@email.com", "2023-02-10T15:04:13Z07:00"},
	{"Eve Anderson", "eve@email.com", "2023-02-10T15:04:14Z07:00"},
}

var benchmarkData = make([]User, 100000)

func TestInsert(t *testing.T) {
	cases := []TestCase[InsertParams[User], map[string]IndexState]{
		{
			given: InsertParams[User]{
				Document: testData[0],
				Language: tokenizer.ENGLISH,
			},
			expected: map[string]IndexState{
				"name": {
					length:      2,
					occurrences: 2,
				},
				"email": {
					length:      3,
					occurrences: 3,
				},
			},
		},
		{
			given: InsertParams[User]{
				Document: testData[1],
				Language: tokenizer.ENGLISH,
			},
			expected: map[string]IndexState{
				"name": {
					length:      1,
					occurrences: 1,
				},
				"email": {
					length:      3,
					occurrences: 3,
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			db := New[User](&Config{
				DefaultLanguage: tokenizer.ENGLISH,
				TokenizerConfig: &tokenizer.Config{},
			})

			v, _ := db.Insert(&c.given)

			assert.NotEmpty(t, v.Id)
			assert.Equal(t, c.given.Document, v.Data)

			assert.Equal(t, 1, len(db.documents))
			assert.Equal(t, len(c.expected), len(db.index.indexes))

			for prop, index := range db.index.indexes {
				assert.Equal(t, c.expected[prop].length, index.Len())
				assert.Equal(t, c.expected[prop].occurrences, len(db.index.tokenOccurrences[prop]))
			}
		})
	}
}

func TestInsertBatch(t *testing.T) {
	cases := []TestCase[InsertBatchParams[User], map[string]IndexState]{
		{
			given: InsertBatchParams[User]{
				Documents: testData,
				BatchSize: 3,
				Language:  tokenizer.ENGLISH,
			},
			expected: map[string]IndexState{
				"name": {
					length:      14,
					occurrences: 14,
				},
				"email": {
					length:      12,
					occurrences: 12,
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			db := New[User](&Config{
				DefaultLanguage: tokenizer.ENGLISH,
				TokenizerConfig: &tokenizer.Config{},
			})

			db.InsertBatch(&c.given)

			assert.Equal(t, len(c.given.Documents), len(db.documents))
			assert.Equal(t, len(c.expected), len(db.index.indexes))

			for prop, index := range db.index.indexes {
				assert.Equal(t, c.expected[prop].length, index.Len())
				assert.Equal(t, c.expected[prop].occurrences, len(db.index.tokenOccurrences[prop]))
			}
		})
	}
}

func TestSearch(t *testing.T) {
	db := New[User](&Config{
		DefaultLanguage: tokenizer.ENGLISH,
		TokenizerConfig: &tokenizer.Config{},
	})
	db.InsertBatch(&InsertBatchParams[User]{
		Documents: testData,
		BatchSize: 3,
		Language:  tokenizer.ENGLISH,
	})

	cases := []TestCase[SearchParams, SearchResult[User]]{
		{
			given: SearchParams{
				Query:      "charlie davis",
				Properties: []string{"name"},
				Offset:     0,
				Limit:      10,
			},
			expected: SearchResult[User]{
				Count: 2,
				Hits: []SearchHit[User]{
					{
						Data:  testData[3],
						Score: 3.4740347056144216,
					},
					{
						Data:  testData[5],
						Score: 1.4816045409242156,
					},
				},
			},
		},
		{
			given: SearchParams{
				Query:      "julia tom",
				Properties: []string{"name", "email"},
				Offset:     0,
				Limit:      10,
			},
			expected: SearchResult[User]{
				Count: 3,
				Hits: []SearchHit[User]{
					{
						Data:  testData[6],
						Score: 5.083472618048522,
					},
					{
						Data:  testData[0],
						Score: 3.4740347056144216,
					},
					{
						Data:  testData[4],
						Score: 1.4816045409242156,
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			actual, _ := db.Search(&c.given)
			log.Println(actual)
			assert.Equal(t, c.expected.Count, actual.Count)

			for i, hit := range c.expected.Hits {
				assert.Equal(t, hit.Data, actual.Hits[i].Data)
				assert.Equal(t, hit.Score, actual.Hits[i].Score)
			}
		})
	}
}

func TestFlattenSchema(t *testing.T) {
	cases := []TestCase[any, map[string]any]{
		{
			given: User{
				Name:   "micpst",
				Email:  "micpst@email.com",
				Joined: "2023-02-10T15:04:05Z07:00",
			},
			expected: map[string]any{
				"name":  "micpst",
				"email": "micpst@email.com",
			},
		},
		{
			given: Document{
				Title:    "The Silicon Brain",
				Abstract: "The human brain is often described as complex and while this is certainly true in many ways, its computational substrate is quite easy to understand.",
				Url:      "https://micpst.com/posts/silicon-brain",
				Author: User{
					Name:   "micpst",
					Email:  "micpst@email.com",
					Joined: "2023-02-10T15:04:05Z07:00",
				},
			},
			expected: map[string]any{
				"title":        "The Silicon Brain",
				"abstract":     "The human brain is often described as complex and while this is certainly true in many ways, its computational substrate is quite easy to understand.",
				"author.name":  "micpst",
				"author.email": "micpst@email.com",
			},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			actual := flattenSchema(c.given)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func BenchmarkInsert(b *testing.B) {
	db := New[User](&Config{
		DefaultLanguage: tokenizer.ENGLISH,
		TokenizerConfig: &tokenizer.Config{},
	})

	for i := 0; i < b.N; i++ {
		for _, data := range benchmarkData {
			_, _ = db.Insert(&InsertParams[User]{
				Document: data,
				Language: tokenizer.ENGLISH,
			})
		}
	}
}

func BenchmarkInsertBatch(b *testing.B) {
	db := New[User](&Config{
		DefaultLanguage: tokenizer.ENGLISH,
		TokenizerConfig: &tokenizer.Config{},
	})

	for i := 0; i < b.N; i++ {
		db.InsertBatch(&InsertBatchParams[User]{
			Documents: benchmarkData,
			BatchSize: 1000,
			Language:  tokenizer.ENGLISH,
		})
	}
}
