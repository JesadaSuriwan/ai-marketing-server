package server

import (
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	"github.com/ai-marketing/ai-marketing-server/pkg/usage/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/usage/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/usage/service"
)

func (s *ginServer) initUsageRouter() {
	usageRepository := repository.NewUsageRepositoryDB(s.db)
	usageService := service.NewUsageService(usageRepository)
	usageHandler := handler.NewUsageHandler(usageService)

	// Exposed so other services can log real token usage/cost per AI call —
	// see server.go for how this gets threaded into promptrun/promptsuggestion/recommendation.
	s.usageService = usageService

	authMiddleware := middlewares.NewAuthMiddleware()
	companyMiddleware := middlewares.NewCompanyMiddleware(s.db)

	router := s.app.Group("/usage")
	router.Use(authMiddleware.AuthRequired)
	router.GET("/summary", companyMiddleware.OwnerByQuery, usageHandler.GetSummary)
}
