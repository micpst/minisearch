package api

import (
	"compress/gzip"
	"encoding/xml"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/micpst/fts-engine/pkg/store"
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
	Query      string     `form:"query" binding:"required"`
	Properties string     `form:"properties"`
	BoolMode   store.Mode `form:"bool_mode"`
}

type UploadDocumentsFileDump struct {
	Documents []Document `xml:"doc"`
}

func (s *Server) uploadDocuments(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	total := 0
	failed := 0
	files := form.File["file[]"]

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

		errs := s.db.InsertBatch(dump.Documents, 10000)
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

func (s *Server) createDocument(c *gin.Context) {
	body := Document{}
	if err := c.BindJSON(&body); err != nil {
		return
	}

	doc, err := s.db.Insert(body)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, documentFromRecord(doc))
}

func (s *Server) updateDocument(c *gin.Context) {
	id := c.Param("id")
	body := Document{}
	if err := c.BindJSON(&body); err != nil {
		return
	}

	doc, err := s.db.Update(id, body)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, documentFromRecord(doc))
}

func (s *Server) deleteDocument(c *gin.Context) {
	id := c.Param("id")
	if err := s.db.Delete(id); err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.Status(http.StatusOK)
}

func (s *Server) searchDocuments(c *gin.Context) {
	params := SearchDocumentsParams{
		Properties: store.WILDCARD,
		BoolMode:   store.AND,
	}
	if err := c.Bind(&params); err != nil {
		return
	}

	start := time.Now()
	docs := s.db.Search(store.SearchParams{
		Query:      params.Query,
		Properties: strings.Split(params.Properties, ","),
		BoolMode:   params.BoolMode,
	})
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