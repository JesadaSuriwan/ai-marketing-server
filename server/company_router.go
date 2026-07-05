package server

import (
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	"github.com/ai-marketing/ai-marketing-server/pkg/company/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/company/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/company/service"
)

func (s *ginServer) initCompanyRouter() {
	companyRepository := repository.NewCompanyRepositoryDB(s.db)
	companyService := service.NewCompanyService(companyRepository)
	companyHandler := handler.NewCompanyHandler(companyService)

	authMiddleware := middlewares.NewAuthMiddleware()
	companyMiddleware := middlewares.NewCompanyMiddleware(s.db)

	router := s.app.Group("/company")
	router.Use(authMiddleware.AuthRequired)

	// List and create do not need ownership middleware (list filters by userId, create sets it)
	router.GET("", companyHandler.GetAll)
	router.POST("", companyHandler.Create)

	// All /:id routes require the caller to own that company
	idGroup := router.Group("/:id")
	idGroup.Use(companyMiddleware.OwnerByParam)
	idGroup.GET("", companyHandler.GetById)
	idGroup.PUT("", companyHandler.Update)
	idGroup.DELETE("", companyHandler.Delete)
	idGroup.POST("/tag", companyHandler.AddTag)
	idGroup.DELETE("/tag/:tag", companyHandler.DeleteTag)
	idGroup.POST("/market", companyHandler.AddMarket)
	idGroup.DELETE("/market/:market", companyHandler.DeleteMarket)
	idGroup.PUT("/social-links", companyHandler.UpdateSocialLinks)
	idGroup.POST("/contact", companyHandler.AddContact)
	idGroup.DELETE("/contact/:contact_id", companyHandler.DeleteContact)
}
