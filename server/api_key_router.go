package server

import (
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	"github.com/ai-marketing/ai-marketing-server/pkg/api_key/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/api_key/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/api_key/service"
)

func (s *ginServer) initApiKeyRouter() {
	apiKeyRepository := repository.NewApiKeyRepositoryDB(s.db)
	apiKeyService := service.NewApiKeyService(apiKeyRepository)
	apiKeyHandler := handler.NewApiKeyHandler(apiKeyService)

	authMiddleware := middlewares.NewAuthMiddleware()
	companyMiddleware := middlewares.NewCompanyMiddleware(s.db)

	router := s.app.Group("/api-key")
	router.Use(authMiddleware.AuthRequired)

	// List: company_id in query → OwnerByQuery
	router.GET("", companyMiddleware.OwnerByQuery, apiKeyHandler.GetAll)
	// Create: company_id in body → OwnerByBodyCompanyId; handler re-reads with ShouldBindBodyWith
	router.POST("", companyMiddleware.OwnerByBodyCompanyId, apiKeyHandler.Create)
	// Revoke: ownership verified inside service via BelongsToUser
	router.DELETE("/:id", apiKeyHandler.Revoke)
}
