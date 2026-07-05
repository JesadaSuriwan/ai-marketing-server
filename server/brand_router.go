package server

import (
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	"github.com/ai-marketing/ai-marketing-server/pkg/brand/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/brand/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/brand/service"
)

func (s *ginServer) initBrandRouter() {
	brandRepository := repository.NewBrandRepositoryDB(s.db)
	brandService := service.NewBrandService(brandRepository)
	brandHandler := handler.NewBrandHandler(brandService)

	authMiddleware := middlewares.NewAuthMiddleware()
	companyMiddleware := middlewares.NewCompanyMiddleware(s.db)

	router := s.app.Group("/brand")
	router.Use(authMiddleware.AuthRequired)

	// List: company_id in query → OwnerByQuery
	router.GET("", companyMiddleware.OwnerByQuery, brandHandler.GetAll)
	// Create: company_id in body → OwnerByBodyCompanyId; handler re-reads body with ShouldBindBodyWith
	router.POST("", companyMiddleware.OwnerByBodyCompanyId, brandHandler.Create)

	// Detail routes: ownership verified inside service via BelongsToUser
	router.GET("/:id", brandHandler.GetById)
	router.PUT("/:id", brandHandler.Update)
	router.DELETE("/:id", brandHandler.Delete)
	router.PUT("/:id/status", brandHandler.UpdateStatus)
}
