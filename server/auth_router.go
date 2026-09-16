package server

import (
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	"github.com/ai-marketing/ai-marketing-server/pkg/auth/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/auth/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/auth/service"
)

func (s *ginServer) initAuthRouter() {
	authRepository := repository.NewAuthRepositoryDB(s.db)
	authService := service.NewAuthService(authRepository)
	authHandler := handler.NewAuthHandler(authService)

	authMiddleware := middlewares.NewAuthMiddleware()
	// 10 burst, 1 request/second steady-state per IP
	rateLimiter := middlewares.NewRateLimitMiddleware(10, 1)

	router := s.app.Group("/auth")
	router.POST("/register", rateLimiter.Limit(), authHandler.Register)
	router.POST("/login", rateLimiter.Limit(), authHandler.Login)

	protected := router.Group("/")
	protected.Use(authMiddleware.AuthRequired)
	protected.GET("/logout", authHandler.Logout)
	protected.GET("/me", authHandler.Me)
	protected.PUT("/change-password", authHandler.ChangePassword)
}
