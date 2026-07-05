package server

import (
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	"github.com/ai-marketing/ai-marketing-server/pkg/notification/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/notification/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/notification/service"
)

func (s *ginServer) initNotificationRouter() {
	notificationRepository := repository.NewNotificationRepositoryDB(s.db)
	notificationService := service.NewNotificationService(notificationRepository)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	authMiddleware := middlewares.NewAuthMiddleware()
	companyMiddleware := middlewares.NewCompanyMiddleware(s.db)

	router := s.app.Group("/notification")
	router.Use(authMiddleware.AuthRequired)

	// List: company_id in query → OwnerByQuery
	router.GET("", companyMiddleware.OwnerByQuery, notificationHandler.GetAll)
	// Create: company_id in body → OwnerByBodyCompanyId; handler re-reads with ShouldBindBodyWith
	router.POST("", companyMiddleware.OwnerByBodyCompanyId, notificationHandler.Create)
	// MarkRead, Delete: ownership verified inside service via BelongsToUser
	router.PUT("/:id/read", notificationHandler.MarkRead)
	// MarkAllRead: company_id in query → OwnerByQuery
	router.PUT("/read-all", companyMiddleware.OwnerByQuery, notificationHandler.MarkAllRead)
	router.DELETE("/:id", notificationHandler.Delete)
}
