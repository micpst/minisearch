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

	d := db.AddDocument(body.Content)
	db.IndexDocument(d)

	c.JSON(http.StatusCreated, DocumentBody{d.Id, d.Content})
}

func UpdateDocument(c *gin.Context) {

}

func DeleteDocument(c *gin.Context) {

}

func SearchDocument(c *gin.Context) {

}
