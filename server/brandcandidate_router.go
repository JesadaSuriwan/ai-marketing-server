package server

import (
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	brandRepository "github.com/ai-marketing/ai-marketing-server/pkg/brand/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/brandcandidate/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/brandcandidate/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/brandcandidate/service"
)

func (s *ginServer) initBrandCandidateRouter() {
	brandCandidateRepository := repository.NewBrandCandidateRepositoryDB(s.db)
	brandRepo := brandRepository.NewBrandRepositoryDB(s.db)
	brandCandidateService := service.NewBrandCandidateService(brandCandidateRepository, brandRepo)
	brandCandidateHandler := handler.NewBrandCandidateHandler(brandCandidateService)

	authMiddleware := middlewares.NewAuthMiddleware()
	companyMiddleware := middlewares.NewCompanyMiddleware(s.db)

	router := s.app.Group("/brand-candidates")
	router.Use(authMiddleware.AuthRequired)

	router.GET("", companyMiddleware.OwnerByQuery, brandCandidateHandler.List)
	// Ownership verified inside the service (candidate.CompanyId vs body company_id).
	router.POST("/:id/resolve", brandCandidateHandler.Resolve)
}
