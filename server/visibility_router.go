package server

import (
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	"github.com/ai-marketing/ai-marketing-server/pkg/visibility/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/visibility/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/visibility/service"
)

func (s *ginServer) initVisibilityRouter() {
	visibilityRepository := repository.NewVisibilityRepositoryDB(s.db)
	visibilityService := service.NewVisibilityService(visibilityRepository)
	visibilityHandler := handler.NewVisibilityHandler(visibilityService)

	authMiddleware := middlewares.NewAuthMiddleware()
	companyMiddleware := middlewares.NewCompanyMiddleware(s.db)

	router := s.app.Group("/visibility")
	router.Use(authMiddleware.AuthRequired)

	// List and breakdown: brand_id in query → OwnerByBrandId
	router.GET("", companyMiddleware.OwnerByBrandId, visibilityHandler.GetAll)
	router.GET("/breakdown", companyMiddleware.OwnerByBrandId, visibilityHandler.GetBreakdown)
	// Create: brand_id in body → OwnerByBodyBrandId; handler re-reads with ShouldBindBodyWith
	router.POST("", companyMiddleware.OwnerByBodyBrandId, visibilityHandler.Create)
}
