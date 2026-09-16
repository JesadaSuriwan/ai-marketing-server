package server

import (
	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	categoryRepository "github.com/ai-marketing/ai-marketing-server/pkg/category/repository"
	companyRepository "github.com/ai-marketing/ai-marketing-server/pkg/company/repository"
	dashboardRepository "github.com/ai-marketing/ai-marketing-server/pkg/dashboard/repository"
	promptRepository "github.com/ai-marketing/ai-marketing-server/pkg/prompt/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/promptsuggestion/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/promptsuggestion/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/promptsuggestion/service"
	"github.com/ai-marketing/ai-marketing-server/providers/anthropic"
)

func (s *ginServer) initPromptSuggestionRouter() {
	if s.conf.Env.AnthropicAPIKey == "" {
		logs.Info("ANTHROPIC_API_KEY not configured: POST /prompt-suggestions/generate will fail until it is set")
	}
	anthropicClient := anthropic.NewClient(s.conf.Env.AnthropicAPIKey)

	suggestionRepository := repository.NewPromptSuggestionRepositoryDB(s.db)
	dashboardRepo := dashboardRepository.NewDashboardRepositoryDB(s.db)
	promptRepo := promptRepository.NewPromptRepositoryDB(s.db)
	categoryRepo := categoryRepository.NewCategoryRepositoryDB(s.db)
	companyRepo := companyRepository.NewCompanyRepositoryDB(s.db)

	suggestionService := service.NewPromptSuggestionService(
		suggestionRepository, dashboardRepo, promptRepo, categoryRepo, companyRepo, s.promptRunService, anthropicClient, s.usageService,
	)
	suggestionHandler := handler.NewPromptSuggestionHandler(suggestionService)

	authMiddleware := middlewares.NewAuthMiddleware()
	companyMiddleware := middlewares.NewCompanyMiddleware(s.db)

	router := s.app.Group("/prompt-suggestions")
	router.Use(authMiddleware.AuthRequired)

	router.GET("", companyMiddleware.OwnerByQuery, suggestionHandler.List)
	router.POST("/generate", companyMiddleware.OwnerByQuery, suggestionHandler.Generate)

	// Detail route: ownership verified inside service via BelongsToUser
	router.PUT("/:id/status", suggestionHandler.UpdateStatus)
}
