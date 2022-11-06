package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Document struct {
	Id       string `json:"id"`
	Title    string `json:"title" xml:"title" binding:"required" index:"true"`
	Url      string `json:"url" xml:"url" binding:"required"`
	Abstract string `json:"abstract" xml:"abstract" binding:"required" index:"true"`
}

type SearchParams struct {
	Query string `form:"q" binding:"required"`
}

func (a *App) createDocument(c *gin.Context) {
	body := Document{}
	if err := c.BindJSON(&body); err != nil {
		return
	}

	_, err := a.db.Create(body)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusCreated)
}

func (a *App) updateDocument(c *gin.Context) {
	id := c.Param("id")
	body := Document{}
	if err := c.BindJSON(&body); err != nil {
		return
	}

	err := a.db.Update(id, body)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.Status(http.StatusOK)
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
	params := SearchParams{}
	if err := c.Bind(&params); err != nil {
		return
	}

	docs := a.db.Search(params.Query)

	c.JSON(http.StatusOK, docs)
}
