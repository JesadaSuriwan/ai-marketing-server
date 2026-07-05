package server

import (
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	"github.com/ai-marketing/ai-marketing-server/pkg/member/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/member/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/member/service"
)

func (s *ginServer) initMemberRouter() {
	memberRepository := repository.NewMemberRepositoryDB(s.db)
	memberService := service.NewMemberService(memberRepository)
	memberHandler := handler.NewMemberHandler(memberService)

	authMiddleware := middlewares.NewAuthMiddleware()
	companyMiddleware := middlewares.NewCompanyMiddleware(s.db)

	router := s.app.Group("/member")
	router.Use(authMiddleware.AuthRequired)

	// List: company_id in query → OwnerByQuery
	router.GET("", companyMiddleware.OwnerByQuery, memberHandler.GetAll)
	// Create: company_id in body → OwnerByBodyCompanyId; handler re-reads with ShouldBindBodyWith
	router.POST("", companyMiddleware.OwnerByBodyCompanyId, memberHandler.Create)
	// Delete: ownership verified inside service via BelongsToUser
	router.DELETE("/:id", memberHandler.Delete)
}
