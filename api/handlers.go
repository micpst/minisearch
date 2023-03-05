package api

import (
	"compress/gzip"
	"encoding/xml"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
	"unsafe"

	"github.com/gin-gonic/gin"
	"github.com/micpst/minisearch/pkg/store"
	"github.com/micpst/minisearch/pkg/tokenizer"
)

type DocumentResponse struct {
	Id       string `json:"id"`
	Title    string `json:"title"`
	Url      string `json:"url"`
	Abstract string `json:"abstract"`
}

type SearchDocument struct {
	Id    string   `json:"id"`
	Data  Document `json:"data"`
	Score float64  `json:"score"`
}

type SearchDocumentResponse struct {
	Count   int              `json:"count"`
	Hits    []SearchDocument `json:"hits"`
	Elapsed int64            `json:"elapsed"`
}

type UploadDocumentsResponse struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Fail    int `json:"fail"`
}

type SearchDocumentsParams struct {
	Query      string             `form:"query" binding:"required"`
	Properties string             `form:"properties"`
	BoolMode   store.Mode         `form:"bool_mode"`
	Offset     int                `form:"offset"`
	Limit      int                `form:"limit"`
	Language   tokenizer.Language `form:"lang"`
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
		dump, err := loadDocumentsFromFile(file)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		params := store.InsertBatchParams[Document]{
			Documents: dump.Documents,
			BatchSize: 10000,
			Language:  tokenizer.Language(strings.ToLower(c.Query("lang"))),
		}
		errs := s.db.InsertBatch(&params)

		total += len(dump.Documents)
		failed += len(errs)
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

	params := store.InsertParams[Document]{
		Document: body,
		Language: tokenizer.Language(strings.ToLower(c.Query("lang"))),
	}

	doc, err := s.db.Insert(&params)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, documentFromRecord(doc))
}

func (s *Server) updateDocument(c *gin.Context) {
	body := Document{}
	if err := c.BindJSON(&body); err != nil {
		return
	}

	params := store.UpdateParams[Document]{
		Id:       c.Param("id"),
		Document: body,
		Language: tokenizer.Language(strings.ToLower(c.Query("lang"))),
	}

	doc, err := s.db.Update(&params)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, documentFromRecord(doc))
}

func (s *Server) deleteDocument(c *gin.Context) {
	params := store.DeleteParams[Document]{
		Id:       c.Param("id"),
		Language: tokenizer.Language(strings.ToLower(c.Query("lang"))),
	}

	if err := s.db.Delete(&params); err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.Status(http.StatusOK)
}

func (s *Server) searchDocuments(c *gin.Context) {
	params := SearchDocumentsParams{
		Properties: store.WILDCARD,
		BoolMode:   store.AND,
		Offset:     0,
		Limit:      10,
	}
	if err := c.Bind(&params); err != nil {
		return
	}

	start := time.Now()
	result, err := s.db.Search(&store.SearchParams{
		Query:      params.Query,
		Properties: strings.Split(params.Properties, ","),
		BoolMode:   params.BoolMode,
		Offset:     params.Offset,
		Limit:      params.Limit,
		Language:   params.Language,
	})
	elapsed := time.Since(start)

	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, SearchDocumentResponse{
		Count:   result.Count,
		Hits:    *(*[]SearchDocument)(unsafe.Pointer(&result.Hits)),
		Elapsed: elapsed.Microseconds(),
	})
}

func documentFromRecord(d store.Record[Document]) DocumentResponse {
	return DocumentResponse{
		Id:       d.Id,
		Title:    d.Data.Title,
		Url:      d.Data.Url,
		Abstract: d.Data.Abstract,
	}
}

func loadDocumentsFromFile(file *multipart.FileHeader) (UploadDocumentsFileDump, error) {
	f, err := file.Open()
	defer func(f multipart.File) {
		_ = f.Close()
	}(f)
	if err != nil {
		return UploadDocumentsFileDump{}, err
	}

	gz, err := gzip.NewReader(f)
	defer func(gz *gzip.Reader) {
		_ = gz.Close()
	}(gz)
	if err != nil {
		return UploadDocumentsFileDump{}, err
	}

	d := xml.NewDecoder(gz)
	dump := UploadDocumentsFileDump{}
	if err := d.Decode(&dump); err != nil {
		return UploadDocumentsFileDump{}, err
	}

	return dump, nil
}
