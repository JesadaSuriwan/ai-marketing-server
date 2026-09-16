package server

import (
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	"github.com/ai-marketing/ai-marketing-server/pkg/company/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/company/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/company/service"
	"github.com/ai-marketing/ai-marketing-server/pkg/permission"
)

func (s *ginServer) initCompanyRouter() {
	companyRepository := repository.NewCompanyRepositoryDB(s.db)
	companyService := service.NewCompanyService(companyRepository, s.db)
	companyHandler := handler.NewCompanyHandler(companyService)

	authMiddleware := middlewares.NewAuthMiddleware()
	companyMiddleware := middlewares.NewCompanyMiddleware(s.db)
	roleMiddleware := middlewares.NewRoleMiddleware(s.db)
	settingsEditors := roleMiddleware.RequireRole(permission.Admin, permission.TeamLead)

	router := s.app.Group("/company")
	router.Use(authMiddleware.AuthRequired)

	// List does not need ownership middleware (filters by userId). Create is
	// gated by CanCreateWorkspace instead of company-ownership, since there's
	// no company yet to check ownership of.
	router.GET("", companyHandler.GetAll)
	router.POST("", roleMiddleware.CanCreateWorkspace, companyHandler.Create)

	// All /:id routes require the caller to own or be an active member of
	// that company; mutating routes are further restricted to Admin/Team
	// Lead — Specialists can't edit any workspace setting (per the role
	// spec, their one settings exception is adding Tags, a separate /category
	// endpoint, not anything here).
	idGroup := router.Group("/:id")
	idGroup.Use(companyMiddleware.OwnerByParam)
	idGroup.GET("", companyHandler.GetById)
	idGroup.PUT("", settingsEditors, companyHandler.Update)
	// Deleting the whole company is owner-only, even though the rest of this
	// group allows accepted team members through.
	idGroup.DELETE("", companyMiddleware.OwnerOnlyByParam, companyHandler.Delete)
	idGroup.POST("/tag", settingsEditors, companyHandler.AddTag)
	idGroup.DELETE("/tag/:tag", settingsEditors, companyHandler.DeleteTag)
	idGroup.POST("/market", settingsEditors, companyHandler.AddMarket)
	idGroup.DELETE("/market/:market", settingsEditors, companyHandler.DeleteMarket)
	idGroup.PUT("/social-links", settingsEditors, companyHandler.UpdateSocialLinks)
	idGroup.POST("/contact", settingsEditors, companyHandler.AddContact)
	idGroup.DELETE("/contact/:contact_id", settingsEditors, companyHandler.DeleteContact)
	idGroup.GET("/ai-engines", companyHandler.GetAiEngines)
	idGroup.PUT("/ai-engines/:platform", settingsEditors, companyHandler.UpdateAiEngine)
	idGroup.GET("/notification-prefs", companyHandler.GetNotificationPrefs)
	idGroup.PUT("/notification-prefs", settingsEditors, companyHandler.UpdateNotificationPrefs)
	idGroup.PUT("/prompt-limit", settingsEditors, companyHandler.UpdatePromptLimit)
	idGroup.PUT("/contract", settingsEditors, companyHandler.UpdateContract)
}
