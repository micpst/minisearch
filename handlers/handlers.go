package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/micpst/full-text-search-engine/db"
)

type DocumentBody struct {
	Id      string `json:"id"`
	Content string `json:"content" binding:"required"`
}

type SearchParams struct {
	Query string `form:"q" binding:"required"`
}

func CreateDocument(c *gin.Context) {
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

func UpdateDocument(c *gin.Context) {
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

func DeleteDocument(c *gin.Context) {
	id := c.Param("id")
	if err := db.RemoveDocument(id); err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.Status(http.StatusOK)
}

func SearchDocument(c *gin.Context) {

}
