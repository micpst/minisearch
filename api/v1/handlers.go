package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/micpst/full-text-search-engine/pkg/db"
)

type DocumentBody struct {
	Id      string `json:"id"`
	Content string `json:"content" binding:"required"`
}

type SearchParams struct {
	Query string `form:"q" binding:"required"`
}

func createDocument(c *gin.Context) {
	var body DocumentBody
	if err := c.BindJSON(&body); err != nil {
		return
	}

	d, err := db.AddDocument(body.Content)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, DocumentBody{d.Id, d.Content})
}

func updateDocument(c *gin.Context) {
	var body DocumentBody
	if err := c.BindJSON(&body); err != nil {
		return
	}

	id := c.Param("id")
	d, err := db.ModifyDocument(id, body.Content)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, DocumentBody{d.Id, d.Content})
}

func deleteDocument(c *gin.Context) {
	id := c.Param("id")
	if err := db.RemoveDocument(id); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusOK)
}

func searchDocument(c *gin.Context) {
	var params SearchParams
	if err := c.Bind(&params); err != nil {
		return
	}
	documents := db.SearchDocuments(params.Query)
	c.JSON(http.StatusOK, documents)
}
