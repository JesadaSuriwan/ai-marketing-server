package server

import (
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	"github.com/ai-marketing/ai-marketing-server/pkg/category/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/category/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/category/service"
)

func (s *ginServer) initCategoryRouter() {
	categoryRepository := repository.NewCategoryRepositoryDB(s.db)
	categoryService := service.NewCategoryService(categoryRepository)
	categoryHandler := handler.NewCategoryHandler(categoryService)

	authMiddleware := middlewares.NewAuthMiddleware()
	companyMiddleware := middlewares.NewCompanyMiddleware(s.db)

	router := s.app.Group("/category")
	router.Use(authMiddleware.AuthRequired)

	router.GET("", companyMiddleware.OwnerByQuery, categoryHandler.GetAll)
	router.POST("", companyMiddleware.OwnerByBodyCompanyId, categoryHandler.Create)

	// Detail routes: ownership verified inside service via BelongsToUser
	router.PUT("/:id", categoryHandler.Update)
	router.DELETE("/:id", categoryHandler.Delete)
}
