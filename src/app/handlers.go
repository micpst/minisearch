package app

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/micpst/full-text-search-engine/src/store"
)

type DocumentResponse struct {
	Id       string `json:"id"`
	Title    string `json:"title"`
	Url      string `json:"url"`
	Abstract string `json:"abstract"`
}

type SearchDocumentResponse struct {
	Count   int                `json:"count"`
	Hits    []DocumentResponse `json:"hits"`
	Elapsed int64              `json:"elapsed"`
}

type SearchDocumentParams struct {
	Query string `form:"q" binding:"required"`
}

func (a *App) createDocument(c *gin.Context) {
	body := Document{}
	if err := c.BindJSON(&body); err != nil {
		return
	}

	doc, err := a.db.Create(body)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, documentFromRecord(doc))
}

func (a *App) updateDocument(c *gin.Context) {
	id := c.Param("id")
	body := Document{}
	if err := c.BindJSON(&body); err != nil {
		return
	}

	doc, err := a.db.Update(id, body)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, documentFromRecord(doc))
}

func (a *App) deleteDocument(c *gin.Context) {
	id := c.Param("id")
	if err := a.db.Delete(id); err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.Status(http.StatusOK)
}

func (a *App) searchDocuments(c *gin.Context) {
	params := SearchDocumentParams{}
	if err := c.Bind(&params); err != nil {
		return
	}

	start := time.Now()
	docs := a.db.Search(params.Query)
	elapsed := time.Since(start)

	c.JSON(http.StatusOK, SearchDocumentResponse{
		Count:   len(docs),
		Hits:    documentListFromRecords(docs),
		Elapsed: elapsed.Microseconds(),
	})
}

func documentFromRecord(d store.Record[Document]) DocumentResponse {
	return DocumentResponse{
		Id:       d.Id,
		Title:    d.S.Title,
		Url:      d.S.Url,
		Abstract: d.S.Abstract,
	}
}

func documentListFromRecords(docs []store.Record[Document]) []DocumentResponse {
	results := make([]DocumentResponse, 0)
	for _, d := range docs {
		doc := documentFromRecord(d)
		results = append(results, doc)
	}
	return results
}
