package server

import (
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	"github.com/ai-marketing/ai-marketing-server/pkg/prompt/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/prompt/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/prompt/service"
)

func (s *ginServer) initPromptRouter() {
	promptRepository := repository.NewPromptRepositoryDB(s.db)
	promptService := service.NewPromptService(promptRepository)
	promptHandler := handler.NewPromptHandler(promptService)

	authMiddleware := middlewares.NewAuthMiddleware()
	companyMiddleware := middlewares.NewCompanyMiddleware(s.db)

	router := s.app.Group("/prompt")
	router.Use(authMiddleware.AuthRequired)

	router.GET("", companyMiddleware.OwnerByQuery, promptHandler.GetAll)
	router.POST("", companyMiddleware.OwnerByBodyCompanyId, promptHandler.Create)

	// Detail routes: ownership verified inside service via BelongsToUser
	router.GET("/:id", promptHandler.GetById)
	router.PUT("/:id", promptHandler.Update)
	router.DELETE("/:id", promptHandler.Delete)
}
