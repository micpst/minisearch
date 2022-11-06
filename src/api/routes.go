package api

func (a *App) initRoutes() {
	a.router.GET("/api/v1/search", a.searchDocuments)
	a.router.POST("/api/v1/documents", a.createDocument)
	a.router.PUT("/api/v1/documents/:id", a.updateDocument)
	a.router.DELETE("/api/v1/documents/:id", a.deleteDocument)
}
