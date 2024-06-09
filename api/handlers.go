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

type SearchRequest struct {
	Query      string             `json:"query" binding:"required"`
	Properties []string           `json:"properties"`
	Exact      bool               `json:"exact"`
	Tolerance  int                `json:"tolerance"`
	Relevance  BM25Params         `json:"relevance"`
	Offset     int                `json:"offset"`
	Limit      int                `json:"limit"`
	Language   tokenizer.Language `json:"lang"`
}

type BM25Params struct {
	K float64 `json:"k"`
	B float64 `json:"b"`
	D float64 `json:"d"`
}

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

type ErrorResponse struct {
	Message string `json:"message"`
}

type UploadDocumentsResponse struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Fail    int `json:"fail"`
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

		errs := s.db.InsertBatch(&store.InsertBatchParams[Document]{
			Documents: dump.Documents,
			BatchSize: 10000,
			Language:  tokenizer.Language(strings.ToLower(c.Query("lang"))),
		})

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

	doc, err := s.db.Insert(&store.InsertParams[Document]{
		Document: body,
		Language: tokenizer.Language(strings.ToLower(c.Query("lang"))),
	})

	switch err.(type) {
	case nil:
		c.JSON(http.StatusCreated, DocumentResponse{
			Id:       doc.Id,
			Title:    doc.Data.Title,
			Url:      doc.Data.Url,
			Abstract: doc.Data.Abstract,
		})
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: err.Error(),
		})
	}
}

func (s *Server) updateDocument(c *gin.Context) {
	body := Document{}
	if err := c.BindJSON(&body); err != nil {
		return
	}

	doc, err := s.db.Update(&store.UpdateParams[Document]{
		Id:       c.Param("id"),
		Document: body,
		Language: tokenizer.Language(strings.ToLower(c.Query("lang"))),
	})

	switch err.(type) {
	case nil:
		c.JSON(http.StatusOK, DocumentResponse{
			Id:       doc.Id,
			Title:    doc.Data.Title,
			Url:      doc.Data.Url,
			Abstract: doc.Data.Abstract,
		})
	case *store.DocumentNotFoundError:
		c.JSON(http.StatusNotFound, ErrorResponse{
			Message: err.Error(),
		})
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: err.Error(),
		})
	}
}

func (s *Server) deleteDocument(c *gin.Context) {
	err := s.db.Delete(&store.DeleteParams[Document]{
		Id:       c.Param("id"),
		Language: tokenizer.Language(strings.ToLower(c.Query("lang"))),
	})

	switch err.(type) {
	case nil:
		c.Status(http.StatusOK)
	case *store.DocumentNotFoundError:
		c.JSON(http.StatusNotFound, ErrorResponse{
			Message: err.Error(),
		})
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: err.Error(),
		})
	}
}

func (s *Server) searchDocuments(c *gin.Context) {
	params := SearchRequest{
		Properties: []string{},
		Limit:      10,
		Relevance: BM25Params{
			K: 1.2,
			B: 0.75,
			D: 0.5,
		},
	}
	if err := c.BindJSON(&params); err != nil {
		return
	}

	start := time.Now()
	result, err := s.db.Search(&store.SearchParams{
		Query:      params.Query,
		Properties: params.Properties,
		Exact:      params.Exact,
		Tolerance:  params.Tolerance,
		Relevance:  store.BM25Params(params.Relevance),
		Offset:     params.Offset,
		Limit:      params.Limit,
	})
	elapsed := time.Since(start)

	switch err.(type) {
	case nil:
		c.JSON(http.StatusOK, SearchDocumentResponse{
			Count:   result.Count,
			Hits:    *(*[]SearchDocument)(unsafe.Pointer(&result.Hits)),
			Elapsed: elapsed.Microseconds(),
		})
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Message: err.Error(),
		})
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
