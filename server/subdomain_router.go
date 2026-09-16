package server

import (
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	"github.com/ai-marketing/ai-marketing-server/pkg/permission"
	"github.com/ai-marketing/ai-marketing-server/pkg/subdomain/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/subdomain/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/subdomain/service"
)

func (s *ginServer) initSubdomainRouter() {
	subdomainRepository := repository.NewSubdomainRepositoryDB(s.db)
	subdomainService := service.NewSubdomainService(subdomainRepository, s.db)
	subdomainHandler := handler.NewSubdomainHandler(subdomainService)

	authMiddleware := middlewares.NewAuthMiddleware()
	companyMiddleware := middlewares.NewCompanyMiddleware(s.db)
	roleMiddleware := middlewares.NewRoleMiddleware(s.db)

	router := s.app.Group("/subdomain")
	router.Use(authMiddleware.AuthRequired)

	// List: company_id in query → OwnerByQuery
	router.GET("", companyMiddleware.OwnerByQuery, subdomainHandler.GetAll)
	// Create: company_id in body → OwnerByBodyCompanyId; Additional Domains is a
	// workspace setting, Admin/Team Lead only.
	router.POST("", companyMiddleware.OwnerByBodyCompanyId, roleMiddleware.RequireRole(permission.Admin, permission.TeamLead), subdomainHandler.Create)
	// Delete: ownership + role verified inside service.
	router.DELETE("/:id", subdomainHandler.Delete)
}
