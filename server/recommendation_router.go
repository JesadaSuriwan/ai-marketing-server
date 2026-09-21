package server

import (
	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	brandRepository "github.com/ai-marketing/ai-marketing-server/pkg/brand/repository"
	citationRepository "github.com/ai-marketing/ai-marketing-server/pkg/citation/repository"
	companyRepository "github.com/ai-marketing/ai-marketing-server/pkg/company/repository"
	dashboardRepository "github.com/ai-marketing/ai-marketing-server/pkg/dashboard/repository"
	promptRepository "github.com/ai-marketing/ai-marketing-server/pkg/prompt/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/recommendation/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/recommendation/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/recommendation/service"
	recommendationRuleRepository "github.com/ai-marketing/ai-marketing-server/pkg/recommendationrule/repository"
	"github.com/ai-marketing/ai-marketing-server/providers/anthropic"
)

func (s *ginServer) initRecommendationRouter() {
	if s.conf.Env.RecommendationClaudeAPIKey == "" {
		logs.Info("RECOMMENDATION_CLAUDE_API_KEY not configured: POST /recommendations/generate will fail until it is set")
	}
	// Deliberately its own key/client, separate from the "claude" prompt-run
	// engine and from citation-extraction's ANTHROPIC_API_KEY, so
	// Recommendations' Claude spend can be tracked independently.
	anthropicClient := anthropic.NewClient(s.conf.Env.RecommendationClaudeAPIKey, s.conf.Env.RecommendationModel)

	recommendationRepo := repository.NewRecommendationRepositoryDB(s.db)
	dashboardRepo := dashboardRepository.NewDashboardRepositoryDB(s.db)
	companyRepo := companyRepository.NewCompanyRepositoryDB(s.db)
	promptRepo := promptRepository.NewPromptRepositoryDB(s.db)
	brandRepo := brandRepository.NewBrandRepositoryDB(s.db)
	citationRepo := citationRepository.NewCitationRepositoryDB(s.db)
	ruleRepo := recommendationRuleRepository.NewRecommendationRuleRepositoryDB(s.db)

	recommendationService := service.NewRecommendationService(recommendationRepo, dashboardRepo, companyRepo, promptRepo, brandRepo, citationRepo, ruleRepo, anthropicClient, s.usageService)
	recommendationHandler := handler.NewRecommendationHandler(recommendationService)

	authMiddleware := middlewares.NewAuthMiddleware()
	companyMiddleware := middlewares.NewCompanyMiddleware(s.db)

	router := s.app.Group("/recommendations")
	router.Use(authMiddleware.AuthRequired)

	router.GET("", companyMiddleware.OwnerByQuery, recommendationHandler.List)
	router.POST("/generate", companyMiddleware.OwnerByQuery, recommendationHandler.Generate)

	// Detail route: ownership verified inside service via BelongsToUser
	router.PUT("/:id/status", recommendationHandler.UpdateStatus)
}
