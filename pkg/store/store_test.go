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
	{"name_01", "email_01@email.com", "2023-02-10T15:04:05Z07:00"},
	{"name_02", "email_02@email.com", "2023-02-10T15:04:06Z07:00"},
	{"name_03", "email_03@email.com", "2023-02-10T15:04:07Z07:00"},
	{"name_04", "email_04@email.com", "2023-02-10T15:04:08Z07:00"},
	{"name_05", "email_05@email.com", "2023-02-10T15:04:09Z07:00"},
	{"name_06", "email_06@email.com", "2023-02-10T15:04:10Z07:00"},
	{"name_07", "email_07@email.com", "2023-02-10T15:04:11Z07:00"},
	{"name_08", "email_08@email.com", "2023-02-10T15:04:12Z07:00"},
	{"name_09", "email_09@email.com", "2023-02-10T15:04:13Z07:00"},
	{"name_10", "email_10@email.com", "2023-02-10T15:04:14Z07:00"},
}

var benchmarkData = make([]User, 100000)

func TestInsert(t *testing.T) {
	db := New[User]()
	data := testData[0]

	v, _ := db.Insert(data)

	assert.NotEmpty(t, v.Id)
	assert.Equal(t, data, v.Data)

	assert.Equal(t, 1, len(db.docs))
	assert.Equal(t, 2, len(db.indexes))

	for _, index := range db.indexes {
		assert.Equal(t, 1, len(index.index))
	}
}

func TestInsertBatch(t *testing.T) {
	db := New[User]()

	errs := db.InsertBatch(testData, 3)
	numInserted := len(testData) - len(errs)

	assert.Equal(t, numInserted, len(db.docs))
	assert.Equal(t, 2, len(db.indexes))

	for _, index := range db.indexes {
		assert.Equal(t, numInserted, len(index.index))
	}
}

func TestSearch(t *testing.T) {
	cases := []TestCase[SearchParams, []User]{
		{
			given: SearchParams{
				Query:      "name_01",
				Properties: []string{"name"},
				BoolMode:   AND,
			},
			expected: []User{
				{"name_01", "email_01@email.com", "2023-02-10T15:04:05Z07:00"},
			},
		},
		{
			given: SearchParams{
				Query:      "name_01 name_02",
				Properties: []string{"name"},
				BoolMode:   OR,
			},
			expected: []User{
				{"name_01", "email_01@email.com", "2023-02-10T15:04:05Z07:00"},
				{"name_02", "email_02@email.com", "2023-02-10T15:04:06Z07:00"},
			},
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			db := New[User]()
			db.InsertBatch(testData, 10)

			actual := db.Search(c.given)

			assert.Equal(t, len(c.expected), len(actual))
			for i := range c.expected {
				assert.Equal(t, c.expected[i], actual[i].Data)
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
