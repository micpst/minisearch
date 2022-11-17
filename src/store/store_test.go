package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestSchema struct {
	NonIndexField string `index:"-"`
	IndexField    string `index:"true"`
}

func TestInsert(t *testing.T) {
	db := New[TestSchema]()
	data := TestSchema{"test_01", "test_01"}

	v, err := db.Insert(data)

	if err != nil {
		assert.Equal(t, 0, db.docs.Len())
		assert.Equal(t, 0, db.index.Len())
		assert.Empty(t, v.Id)
		assert.Equal(t, TestSchema{}, v.S)
	} else {
		assert.Equal(t, 1, db.docs.Len())
		assert.Equal(t, 1, db.index.Len())
		assert.NotEmpty(t, v.Id)
		assert.Equal(t, data, v.S)
	}
}

func TestInsertBatch(t *testing.T) {
	db := New[TestSchema]()
	data := []TestSchema{
		{"test_01", "test_01"},
		{"test_02", "test_02"},
		{"test_03", "test_03"},
		{"test_04", "test_04"},
		{"test_05", "test_05"},
		{"test_06", "test_06"},
		{"test_07", "test_07"},
		{"test_08", "test_08"},
		{"test_09", "test_09"},
		{"test_10", "test_10"},
	}

	errs := db.InsertBatch(data, 10)

	assert.Equal(t, len(data)-len(errs), db.docs.Len())
	assert.Equal(t, len(data)-len(errs), db.index.Len())
}
