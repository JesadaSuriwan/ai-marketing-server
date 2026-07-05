package server

import (
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	"github.com/ai-marketing/ai-marketing-server/pkg/citation/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/citation/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/citation/service"
)

func (s *ginServer) initCitationRouter() {
	citationRepository := repository.NewCitationRepositoryDB(s.db)
	citationService := service.NewCitationService(citationRepository)
	citationHandler := handler.NewCitationHandler(citationService)

	authMiddleware := middlewares.NewAuthMiddleware()
	companyMiddleware := middlewares.NewCompanyMiddleware(s.db)

	router := s.app.Group("/citation")
	router.Use(authMiddleware.AuthRequired)

	// List: prompt_id in query → OwnerByPromptId
	router.GET("", companyMiddleware.OwnerByPromptId, citationHandler.GetAll)
	// Create: prompt_id in body → OwnerByBodyPromptId; handler re-reads with ShouldBindBodyWith
	router.POST("", companyMiddleware.OwnerByBodyPromptId, citationHandler.Create)

	// Detail routes: ownership verified inside service via BelongsToUser
	router.GET("/:id", citationHandler.GetById)
	router.PUT("/:id", citationHandler.Update)
	router.DELETE("/:id", citationHandler.Delete)
	router.PUT("/:id/notes", citationHandler.UpdateNotes)
	router.PUT("/:id/archive", citationHandler.UpdateArchive)
}
