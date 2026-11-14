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

	// List: shared inbox across every company the user can access — scoped
	// to userId (from AuthRequired) inside the handler, no company_id needed.
	router.GET("", notificationHandler.GetAll)
	// Create: company_id in body → OwnerByBodyCompanyId; handler re-reads with ShouldBindBodyWith
	router.POST("", companyMiddleware.OwnerByBodyCompanyId, notificationHandler.Create)
	// MarkRead, Delete: ownership verified inside service via BelongsToUser
	router.PUT("/:id/read", notificationHandler.MarkRead)
	// MarkAllRead: same shared-inbox scope as GetAll, by userId
	router.PUT("/read-all", notificationHandler.MarkAllRead)
	router.DELETE("/:id", notificationHandler.Delete)
}
