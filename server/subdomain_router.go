package server

import (
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	"github.com/ai-marketing/ai-marketing-server/pkg/subdomain/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/subdomain/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/subdomain/service"
)

func (s *ginServer) initSubdomainRouter() {
	subdomainRepository := repository.NewSubdomainRepositoryDB(s.db)
	subdomainService := service.NewSubdomainService(subdomainRepository)
	subdomainHandler := handler.NewSubdomainHandler(subdomainService)

	authMiddleware := middlewares.NewAuthMiddleware()
	companyMiddleware := middlewares.NewCompanyMiddleware(s.db)

	router := s.app.Group("/subdomain")
	router.Use(authMiddleware.AuthRequired)

	// List: company_id in query → OwnerByQuery
	router.GET("", companyMiddleware.OwnerByQuery, subdomainHandler.GetAll)
	// Create: company_id in body → OwnerByBodyCompanyId; handler re-reads with ShouldBindBodyWith
	router.POST("", companyMiddleware.OwnerByBodyCompanyId, subdomainHandler.Create)
	// Delete: ownership verified inside service via BelongsToUser
	router.DELETE("/:id", subdomainHandler.Delete)
}
