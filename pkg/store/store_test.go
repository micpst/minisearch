package store

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestCase[Given any, Expected any] struct {
	given    Given
	expected Expected
}

type IndexState struct {
	length      int
	occurrences int
	docsCount   int
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
	cases := []TestCase[User, map[string]IndexState]{
		{
			given: testData[0],
			expected: map[string]IndexState{
				"name": {
					length:      2,
					occurrences: 2,
					docsCount:   1,
				},
				"email": {
					length:      1,
					occurrences: 1,
					docsCount:   1,
				},
			},
		},
		{
			given: testData[1],
			expected: map[string]IndexState{
				"name": {
					length:      1,
					occurrences: 1,
					docsCount:   1,
				},
				"email": {
					length:      1,
					occurrences: 1,
					docsCount:   1,
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			db := New[User]()

			v, _ := db.Insert(c.given)

			assert.NotEmpty(t, v.Id)
			assert.Equal(t, c.given, v.Data)

			assert.Equal(t, 1, len(db.docs))
			assert.Equal(t, 2, len(db.indexes))

			for prop, index := range db.indexes {
				assert.Equal(t, c.expected[prop].length, len(index.index))
				assert.Equal(t, c.expected[prop].occurrences, len(index.occurrences))
				assert.Equal(t, c.expected[prop].docsCount, index.docsCount)
			}
		})
	}
}

func TestInsertBatch(t *testing.T) {
	cases := []TestCase[[]User, map[string]IndexState]{
		{
			given: testData,
			expected: map[string]IndexState{
				"name": {
					length:      14,
					occurrences: 14,
					docsCount:   10,
				},
				"email": {
					length:      10,
					occurrences: 10,
					docsCount:   10,
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			db := New[User]()

			db.InsertBatch(c.given, 3)

			assert.Equal(t, 10, len(db.docs))
			assert.Equal(t, 2, len(db.indexes))

			for prop, index := range db.indexes {
				assert.Equal(t, c.expected[prop].length, len(index.index))
				assert.Equal(t, c.expected[prop].occurrences, len(index.occurrences))
				assert.Equal(t, c.expected[prop].docsCount, index.docsCount)
			}
		})
	}
}

func TestSearch(t *testing.T) {
	db := New[User]()
	db.InsertBatch(testData, len(testData))

	cases := []TestCase[SearchParams, []SearchResult[User]]{
		{
			given: SearchParams{
				Query:      "charlie davis",
				Properties: []string{"name"},
				BoolMode:   AND,
			},
			expected: []SearchResult[User]{
				{
					Data:  testData[3],
					Score: 1.7370173528072108,
				},
			},
		},
		{
			given: SearchParams{
				Query:      "julia tom@email.com",
				Properties: []string{"name", "email"},
				BoolMode:   OR,
			},
			expected: []SearchResult[User]{
				{
					Data:  testData[0],
					Score: 1.992430164690206,
				},
				{
					Data:  testData[6],
					Score: 0.996215082345103,
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			actual := db.Search(c.given)

			assert.Equal(t, len(c.expected), len(actual))

			for i := range c.expected {
				assert.Equal(t, c.expected[i].Data, actual[i].Data)
				assert.Equal(t, c.expected[i].Score, actual[i].Score)
			}
		})
	}
}

func TestFlattenSchema(t *testing.T) {
	cases := []TestCase[any, map[string]string]{
		{
			given: User{
				Name:   "micpst",
				Email:  "micpst@email.com",
				Joined: "2023-02-10T15:04:05Z07:00",
			},
			expected: map[string]string{
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
			expected: map[string]string{
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
	db := New[User]()

	for i := 0; i < b.N; i++ {
		for _, data := range benchmarkData {
			_, _ = db.Insert(data)
		}
	}
}

func BenchmarkInsertBatch(b *testing.B) {
	db := New[User]()

	for i := 0; i < b.N; i++ {
		db.InsertBatch(benchmarkData, 1000)
	}
}
