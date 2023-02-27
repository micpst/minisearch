package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/micpst/minisearch/pkg/store"
)

type Config struct {
	Port               uint
	MaxMultipartMemory int64
}

type Server struct {
	config *Config
	db     *store.MemDB[Document]
	router *gin.Engine
}

func New(c *Config) *Server {
	s := &Server{
		config: c,
		db:     store.New[Document](),
		router: gin.Default(),
	}
	s.router.MaxMultipartMemory = s.config.MaxMultipartMemory
	s.initRoutes()
	return s
}

func (s *Server) initRoutes() {
	s.router.GET("/api/v1/search", s.searchDocuments)
	s.router.POST("/api/v1/upload", s.uploadDocuments)
	s.router.POST("/api/v1/documents", s.createDocument)
	s.router.PUT("/api/v1/documents/:id", s.updateDocument)
	s.router.DELETE("/api/v1/documents/:id", s.deleteDocument)
}

func (s *Server) Run() error {
	return s.router.Run(fmt.Sprintf(":%d", s.config.Port))
}
