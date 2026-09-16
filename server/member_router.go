package server

import (
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	authRepository "github.com/ai-marketing/ai-marketing-server/pkg/auth/repository"
	companyRepository "github.com/ai-marketing/ai-marketing-server/pkg/company/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/member/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/member/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/member/service"
	"github.com/ai-marketing/ai-marketing-server/pkg/permission"
)

func (s *ginServer) initMemberRouter() {
	memberRepository := repository.NewMemberRepositoryDB(s.db)
	authRepo := authRepository.NewAuthRepositoryDB(s.db)
	companyRepo := companyRepository.NewCompanyRepositoryDB(s.db)
	memberService := service.NewMemberService(memberRepository, authRepo, companyRepo, s.db)
	memberHandler := handler.NewMemberHandler(memberService)

	authMiddleware := middlewares.NewAuthMiddleware()
	companyMiddleware := middlewares.NewCompanyMiddleware(s.db)
	roleMiddleware := middlewares.NewRoleMiddleware(s.db)

	router := s.app.Group("/member")

	authed := router.Group("")
	authed.Use(authMiddleware.AuthRequired)

	// List: company_id in query → OwnerByQuery
	authed.GET("", companyMiddleware.OwnerByQuery, memberHandler.GetAll)
	// Create: company_id in body → OwnerByBodyCompanyId; only Admin/Team Lead invite new members.
	authed.POST("", companyMiddleware.OwnerByBodyCompanyId, roleMiddleware.RequireRole(permission.Admin, permission.TeamLead), memberHandler.Create)
	// Delete: the removal permission matrix (who-can-remove-whom) is checked inside the service.
	authed.DELETE("/:id", memberHandler.Delete)
}
