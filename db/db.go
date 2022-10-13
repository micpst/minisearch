package db

import (
	"github.com/cornelk/hashmap"
	"github.com/gofrs/uuid"
)

var (
	documents *hashmap.Map[string, string]
	//indexes   *hashmap.Map[string, int]
)

func init() {
	documents = hashmap.New[string, string]()
	//indexes = hashmap.New[string, int]()
}

func AddDocument(data string) string {
	id := uuid.Must(uuid.NewV4()).String()
	documents.Insert(id, data)
	return id
}

func IndexDocument(data string) {
	// TODO Implement IndexDocument()
}
