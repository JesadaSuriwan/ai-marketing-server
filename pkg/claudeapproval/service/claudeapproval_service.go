package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	authRepository "github.com/ai-marketing/ai-marketing-server/pkg/auth/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/claudeapproval/repository"
	companyRepository "github.com/ai-marketing/ai-marketing-server/pkg/company/repository"
	memberRepository "github.com/ai-marketing/ai-marketing-server/pkg/member/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/permission"
)

const approvalTokenTTL = 7 * 24 * time.Hour

type claudeApprovalService struct {
	claudeApprovalRepository repository.ClaudeApprovalRepository
	companyRepository        companyRepository.CompanyRepository
	authRepository           authRepository.AuthRepository
	memberRepository         memberRepository.MemberRepository
	emailProvider            EmailProvider
	frontendBaseURL          string
}

func NewClaudeApprovalService(
	claudeApprovalRepository repository.ClaudeApprovalRepository,
	companyRepository companyRepository.CompanyRepository,
	authRepository authRepository.AuthRepository,
	memberRepository memberRepository.MemberRepository,
	emailProvider EmailProvider,
	frontendBaseURL string,
) ClaudeApprovalService {
	return claudeApprovalService{
		claudeApprovalRepository, companyRepository, authRepository, memberRepository, emailProvider, frontendBaseURL,
	}
}

func generateApprovalToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func requestEmailHTML(companyName, requesterName, link string) string {
	return fmt.Sprintf(`
		<div style="font-family: sans-serif; max-width: 480px; margin: 0 auto;">
			<h2>Claude access requested</h2>
			<p><strong>%s</strong> is requesting to turn on the Claude AI agent for <strong>%s</strong>.</p>
			<p><a href="%s" style="display:inline-block;background:#9B5DE5;color:#fff;padding:10px 20px;border-radius:8px;text-decoration:none;font-weight:600;">Approve Request</a></p>
			<p style="color:#888;font-size:12px;">This link expires in 7 days. If you don't recognize this request, you can ignore this email.</p>
		</div>
	`, requesterName, companyName, link)
}

func approvedEmailHTML(companyName string) string {
	return fmt.Sprintf(`
		<div style="font-family: sans-serif; max-width: 480px; margin: 0 auto;">
			<h2>Claude approved</h2>
			<p>Your request to turn on the Claude AI agent for <strong>%s</strong> has been approved. It's live now.</p>
		</div>
	`, companyName)
}

// Create is called by a Specialist requesting Claude access — recipients are
// the workspace creator plus any active Team Lead members, matching the
// role that's actually allowed to approve it.
func (s claudeApprovalService) Create(companyId, requestedByUserId int) (*SimpleResponse, error) {
	company, err := s.companyRepository.GetById(companyId)
	if err != nil {
		return nil, errs.NewNotFoundError("company not found")
	}

	requester, err := s.authRepository.GetUserById(requestedByUserId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	recipients := map[string]bool{}
	if creator, err := s.authRepository.GetUserById(company.UserId); err == nil && creator != nil {
		recipients[creator.Email] = true
	}
	if members, err := s.memberRepository.GetAll(companyId); err == nil {
		for _, m := range members {
			if m.Role == permission.TeamLead && m.Status == "active" {
				recipients[m.Email] = true
			}
		}
	}

	token, err := generateApprovalToken()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	expiresAt := time.Now().Add(approvalTokenTTL)

	tx, err := s.claudeApprovalRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if _, err = s.claudeApprovalRepository.Create(tx, repository.ClaudeApprovalRequest{
		CompanyId: companyId, RequestedByUserId: requestedByUserId, Status: "pending",
		Token: token, TokenExpiresAt: expiresAt,
	}); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	// Best-effort per recipient — the request is already safely recorded
	// regardless of whether any individual email goes out.
	link := fmt.Sprintf("%s/claude-approval?token=%s", strings.TrimRight(s.frontendBaseURL, "/"), token)
	for email := range recipients {
		if err := s.emailProvider.Send(email, fmt.Sprintf("Claude access requested for %s", company.Name), requestEmailHTML(company.Name, requester.Name, link)); err != nil {
			logs.Error(fmt.Errorf("failed to send claude approval request email to %s: %w", email, err))
		}
	}

	return &SimpleResponse{Status: true, Desc: "Request sent to your workspace's Team Lead"}, nil
}

func (s claudeApprovalService) Approve(token string) (*SimpleResponse, error) {
	req, err := s.claudeApprovalRepository.GetByToken(token)
	if err != nil || req.Status != "pending" {
		return nil, errs.NewBadRequestError("invalid or already-used approval link")
	}
	if time.Now().After(req.TokenExpiresAt) {
		return nil, errs.NewBadRequestError("this approval link has expired")
	}

	tx, err := s.claudeApprovalRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.companyRepository.SetAiEngine(tx, req.CompanyId, "claude", true); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	if err = s.claudeApprovalRepository.MarkResolved(tx, req.Id, "approved"); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	companyName := "your workspace"
	if company, err := s.companyRepository.GetById(req.CompanyId); err == nil {
		companyName = company.Name
	}
	if requester, err := s.authRepository.GetUserById(req.RequestedByUserId); err == nil {
		if err := s.emailProvider.Send(requester.Email, fmt.Sprintf("Claude approved for %s", companyName), approvedEmailHTML(companyName)); err != nil {
			logs.Error(fmt.Errorf("failed to send claude approval confirmation to %s: %w", requester.Email, err))
		}
	}

	return &SimpleResponse{Status: true, Desc: "Claude has been enabled"}, nil
}
