package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/micpst/full-text-search-engine/src/storage"
)

type App struct {
	db     *storage.MemDB
	router *gin.Engine
}

func New() *App {
	app := &App{
		db:     storage.New(),
		router: gin.Default(),
	}
	app.initRoutes()
	return app
}

func (a *App) Run(port *uint) error {
	return a.router.Run(fmt.Sprintf(":%d", *port))
}
