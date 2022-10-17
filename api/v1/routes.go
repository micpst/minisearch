package v1

import (
	"github.com/gin-gonic/gin"
)

func InitRoutes() *gin.Engine {
	r := gin.Default()

	r.GET("/api/v1/search", searchDocument)
	r.POST("/api/v1/documents", createDocument)
	r.PUT("/api/v1/documents/:id", updateDocument)
	r.DELETE("/api/v1/documents/:id", deleteDocument)

	return r
}
