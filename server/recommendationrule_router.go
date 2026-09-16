package server

import (
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	"github.com/ai-marketing/ai-marketing-server/pkg/recommendationrule/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/recommendationrule/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/recommendationrule/service"
)

func (s *ginServer) initRecommendationRuleRouter() {
	recommendationRuleRepository := repository.NewRecommendationRuleRepositoryDB(s.db)
	recommendationRuleService := service.NewRecommendationRuleService(recommendationRuleRepository)
	recommendationRuleHandler := handler.NewRecommendationRuleHandler(recommendationRuleService)

	authMiddleware := middlewares.NewAuthMiddleware()
	companyMiddleware := middlewares.NewCompanyMiddleware(s.db)

	router := s.app.Group("/recommendation-rules")
	router.Use(authMiddleware.AuthRequired)

	router.GET("", companyMiddleware.OwnerByQuery, recommendationRuleHandler.GetAll)
	router.POST("", companyMiddleware.OwnerByBodyCompanyId, recommendationRuleHandler.Create)

	// Detail routes: ownership verified inside service via BelongsToUser
	router.PUT("/:id", recommendationRuleHandler.Update)
	router.DELETE("/:id", recommendationRuleHandler.Delete)
}
