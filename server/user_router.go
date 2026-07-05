package server

import (
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	"github.com/ai-marketing/ai-marketing-server/pkg/user/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/user/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/user/service"
)

func (s *ginServer) initUserRouter() {
	userRepository := repository.NewUserRepositoryDB(s.db)
	userService := service.NewUserService(userRepository)
	userHandler := handler.NewUserHandler(userService)

	authMiddleware := middlewares.NewAuthMiddleware()

	router := s.app.Group("/user")
	router.Use(authMiddleware.AuthRequired)
	router.GET("/:id", userHandler.GetById)
	router.PUT("/:id", userHandler.Update)
}
