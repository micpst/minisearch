package app

import (
	"compress/gzip"
	"encoding/xml"
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

type SearchDocumentsResponse struct {
	Count   int                `json:"count"`
	Hits    []DocumentResponse `json:"hits"`
	Elapsed int64              `json:"elapsed"`
}

type UploadDocumentsResponse struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Fail    int `json:"fail"`
}

type SearchDocumentsParams struct {
	Query string `form:"q" binding:"required"`
}

type UploadDocumentsFileDump struct {
	Documents []Document `xml:"doc"`
}

func (a *App) uploadDocuments(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	total := 0
	failed := 0
	files := form.File["file"]

	for _, file := range files {
		f, err := file.Open()
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		gz, err := gzip.NewReader(f)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		d := xml.NewDecoder(gz)
		dump := UploadDocumentsFileDump{}
		if err := d.Decode(&dump); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		errs := a.db.InsertBatch(dump.Documents, 10000)
		total += len(dump.Documents)
		failed += len(errs)

		_ = f.Close()
		_ = gz.Close()
	}

	c.JSON(http.StatusOK, UploadDocumentsResponse{
		Total:   total,
		Success: total - failed,
		Fail:    failed,
	})
}

func (a *App) createDocument(c *gin.Context) {
	body := Document{}
	if err := c.BindJSON(&body); err != nil {
		return
	}

	doc, err := a.db.Insert(body)
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
	params := SearchDocumentsParams{}
	if err := c.Bind(&params); err != nil {
		return
	}

	start := time.Now()
	docs := a.db.Search(params.Query)
	elapsed := time.Since(start)

	c.JSON(http.StatusOK, SearchDocumentsResponse{
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
