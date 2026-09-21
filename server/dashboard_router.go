package server

import (
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	"github.com/ai-marketing/ai-marketing-server/pkg/dashboard/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/dashboard/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/dashboard/service"
)

func (s *ginServer) initDashboardRouter() {
	dashboardRepository := repository.NewDashboardRepositoryDB(s.db)
	dashboardService := service.NewDashboardService(dashboardRepository)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)

	authMiddleware := middlewares.NewAuthMiddleware()
	companyMiddleware := middlewares.NewCompanyMiddleware(s.db)

	router := s.app.Group("/dashboard")
	router.Use(authMiddleware.AuthRequired)
	router.Use(companyMiddleware.OwnerByQuery)

	router.GET("/stats", dashboardHandler.GetStats)
	router.GET("/top-prompts", dashboardHandler.GetTopPrompts)
	router.GET("/top-domains", dashboardHandler.GetTopDomains)
	router.GET("/recent-citations", dashboardHandler.GetRecentCitations)
	router.GET("/visibility-trend", dashboardHandler.GetVisibilityTrend)
	router.GET("/platform-breakdown", dashboardHandler.GetPlatformBreakdown)
	router.GET("/prompt-trend", dashboardHandler.GetPromptTrend)
	router.GET("/company-metrics", dashboardHandler.GetCompanyMetrics)
	router.GET("/prompt-rankings", dashboardHandler.GetPromptRankings)
	router.GET("/prompt-domains", dashboardHandler.GetPromptDomains)
	router.GET("/prompts-overview", dashboardHandler.GetPromptsOverview)
	router.GET("/brand-ranking", dashboardHandler.GetBrandRanking)
	router.GET("/brand-coverage-trend", dashboardHandler.GetBrandCoverageTrend)
	router.GET("/top-prompts-by-brand", dashboardHandler.GetTopPromptsByBrand)
	router.GET("/top-citation-urls", dashboardHandler.GetTopCitationURLs)
	router.GET("/citation-urls", dashboardHandler.GetCitationURLs)
	router.GET("/citation-url-prompts", dashboardHandler.GetCitationURLPrompts)
	router.GET("/citation-winners-losers", dashboardHandler.GetCitationWinnersLosers)
	router.GET("/brand-citations", dashboardHandler.GetBrandCitations)
}
