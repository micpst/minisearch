package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/micpst/full-text-search-engine/src/storage"
)

type SearchParams struct {
	query string `form:"q" binding:"required"`
}

func (a *App) createDocument(c *gin.Context) {
	body := storage.Document{}
	if err := c.BindJSON(&body); err != nil {
		return
	}

	doc, err := a.db.Create(body)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, doc)
}

func (a *App) updateDocument(c *gin.Context) {
	id := c.Param("id")
	body := storage.Document{}
	if err := c.BindJSON(&body); err != nil {
		return
	}

	doc, err := a.db.Update(id, body)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, doc)
}

func (a *App) deleteDocument(c *gin.Context) {
	id := c.Param("id")
	if err := a.db.Delete(id); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Status(http.StatusOK)
}

func (a *App) searchDocument(c *gin.Context) {
	params := SearchParams{}
	if err := c.Bind(&params); err != nil {
		return
	}
	docs := a.db.Search(params.query)
	c.JSON(http.StatusOK, docs)
}
