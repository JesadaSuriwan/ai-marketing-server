package server

import (
	"strings"

	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/middlewares"
	authRepository "github.com/ai-marketing/ai-marketing-server/pkg/auth/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/claudeapproval/handler"
	"github.com/ai-marketing/ai-marketing-server/pkg/claudeapproval/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/claudeapproval/service"
	companyRepository "github.com/ai-marketing/ai-marketing-server/pkg/company/repository"
	memberRepository "github.com/ai-marketing/ai-marketing-server/pkg/member/repository"
	"github.com/ai-marketing/ai-marketing-server/providers/resend"
)

func (s *ginServer) initClaudeApprovalRouter() {
	if s.conf.Env.ResendAPIKey == "" {
		logs.Info("RESEND_API_KEY not configured: Claude approval request emails will fail to send until it is set")
	}
	emailClient := resend.NewClient(s.conf.Env.ResendAPIKey, s.conf.Env.EmailFrom)
	frontendBaseURL := strings.TrimSpace(strings.Split(s.conf.Server.ClientURL, ",")[0])

	claudeApprovalRepo := repository.NewClaudeApprovalRepositoryDB(s.db)
	companyRepo := companyRepository.NewCompanyRepositoryDB(s.db)
	authRepo := authRepository.NewAuthRepositoryDB(s.db)
	memberRepo := memberRepository.NewMemberRepositoryDB(s.db)

	claudeApprovalService := service.NewClaudeApprovalService(
		claudeApprovalRepo, companyRepo, authRepo, memberRepo, emailClient, frontendBaseURL,
	)
	claudeApprovalHandler := handler.NewClaudeApprovalHandler(claudeApprovalService)

	authMiddleware := middlewares.NewAuthMiddleware()
	companyMiddleware := middlewares.NewCompanyMiddleware(s.db)

	router := s.app.Group("/claude-approval")

	// Public — the approving Team Lead clicks this straight from their inbox,
	// same trust model as /member/invite/:token/accept.
	router.GET("/:token/approve", claudeApprovalHandler.Approve)

	// Any active member can request (a Specialist is the expected caller,
	// but nothing stops Admin/Team Lead from using it too — they'd just
	// approve their own request the same click).
	authed := router.Group("")
	authed.Use(authMiddleware.AuthRequired)
	authed.POST("", companyMiddleware.OwnerByBodyCompanyId, claudeApprovalHandler.Create)
}
