package server

import (
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	"github.com/ai-marketing/ai-marketing-server/pkg/branddomain/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/branddomain/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/branddomain/service"
)

func (s *ginServer) initBrandDomainRouter() {
	brandDomainRepository := repository.NewBrandDomainRepositoryDB(s.db)
	brandDomainService := service.NewBrandDomainService(brandDomainRepository, s.db)
	brandDomainHandler := handler.NewBrandDomainHandler(brandDomainService)

	authMiddleware := middlewares.NewAuthMiddleware()
	companyMiddleware := middlewares.NewCompanyMiddleware(s.db)

	router := s.app.Group("/brand-domain")
	router.Use(authMiddleware.AuthRequired)

	// List: brand_id in query → OwnerByBrandId
	router.GET("", companyMiddleware.OwnerByBrandId, brandDomainHandler.GetAll)
	// Create: brand_id in body → OwnerByBodyBrandId; fine-grained role check
	// (Admin/Team Lead/Specialist, not Customer) happens inside the service.
	router.POST("", companyMiddleware.OwnerByBodyBrandId, brandDomainHandler.Create)
	// Delete: ownership + role verified inside service.
	router.DELETE("/:id", brandDomainHandler.Delete)
}
