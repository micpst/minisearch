package main

import (
	"github.com/gin-gonic/gin"
	"github.com/micpst/full-text-search-engine/handlers"
)

func initRoutes(r *gin.Engine) {
	r.GET("/api/v1/search", handlers.SearchDocument)
	r.POST("/api/v1/documents", handlers.CreateDocument)
	r.PUT("/api/v1/documents/:id", handlers.UpdateDocument)
	r.DELETE("/api/v1/documents/:id", handlers.DeleteDocument)
}
