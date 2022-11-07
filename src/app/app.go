package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/micpst/full-text-search-engine/src/store"
)

type App struct {
	db     *store.MemDB[Document]
	router *gin.Engine
}

func New() *App {
	app := &App{
		db:     store.New[Document](),
		router: gin.Default(),
	}
	app.initRoutes()
	return app
}

func (a *App) Run(port *uint) error {
	return a.router.Run(fmt.Sprintf(":%d", *port))
}
