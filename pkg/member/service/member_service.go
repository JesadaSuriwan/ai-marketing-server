package service

import (
	"strings"

	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	authRepository "github.com/ai-marketing/ai-marketing-server/pkg/auth/repository"
	companyRepository "github.com/ai-marketing/ai-marketing-server/pkg/company/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/member/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/permission"
	"github.com/ai-marketing/ai-marketing-server/utils"
	"github.com/jmoiron/sqlx"
)

type memberService struct {
	memberRepository  repository.MemberRepository
	authRepository    authRepository.AuthRepository
	companyRepository companyRepository.CompanyRepository
	db                *sqlx.DB
}

func NewMemberService(
	memberRepository repository.MemberRepository,
	authRepository authRepository.AuthRepository,
	companyRepository companyRepository.CompanyRepository,
	db *sqlx.DB,
) MemberService {
	return memberService{memberRepository, authRepository, companyRepository, db}
}

// canRemove encodes the member-removal permission matrix: Admin can remove
// anyone; Team Lead can remove a Specialist or Customer but not another Team
// Lead; Specialist can remove a Customer only; Customer can't remove anyone.
func canRemove(requesterRole, targetRole string) bool {
	switch requesterRole {
	case permission.Admin:
		return true
	case permission.TeamLead:
		return targetRole == permission.Specialist || targetRole == permission.Customer
	case permission.Specialist:
		return targetRole == permission.Customer
	default:
		return false
	}
}

func (s memberService) GetAll(companyId int) (*MemberListResponse, error) {
	members, err := s.memberRepository.GetAll(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	data := []MemberData{}
	for _, m := range members {
		data = append(data, MemberData{Id: m.Id, CompanyId: m.CompanyId, Email: m.Email, Name: m.Name, Role: m.Role, JoinedAt: m.JoinedAt, Status: m.Status})
	}

	return &MemberListResponse{Status: true, Desc: "Get members successful", Data: data}, nil
}

// Create adds a member directly — no email is sent. If the email already
// has an account, that account is linked immediately (status "active"). If
// not, a new account is created on the spot with a one-time temp password
// (returned only in this response, never persisted in plaintext) and
// MustChangePassword set, forcing the new member to pick their own password
// the moment they log in.
func (s memberService) Create(req CreateMemberRequest) (*CreateMemberResponse, error) {
	if _, err := s.companyRepository.GetById(req.CompanyId); err != nil {
		return nil, errs.NewNotFoundError("company not found")
	}

	role := strings.ToLower(strings.TrimSpace(req.Role))
	if role == "" {
		role = permission.Specialist
	}
	if role != permission.TeamLead && role != permission.Specialist && role != permission.Customer {
		return nil, errs.NewBadRequestError("invalid role")
	}

	existingUser, _ := s.authRepository.GetUserByEmail(req.Email)

	tx, err := s.memberRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	var userId int
	var name string
	tempPassword := ""

	if existingUser != nil {
		userId = existingUser.Id
		name = existingUser.Name
	} else {
		name = strings.TrimSpace(req.Name)
		if name == "" {
			name = req.Email
		}
		initials := strings.ToUpper(name[:min(2, len(name))])

		tempPassword, err = utils.GenerateTempPassword()
		if err != nil {
			logs.Error(err)
			return nil, errs.NewUnexpectedError()
		}

		hashed, err := utils.HashPassword(tempPassword)
		if err != nil {
			logs.Error(err)
			return nil, errs.NewUnexpectedError()
		}

		userId, err = s.authRepository.CreateUser(tx, authRepository.User{
			Email: req.Email, Password: hashed, Name: name, Initials: initials, MustChangePassword: true,
		})
		if err != nil {
			logs.Error(err)
			return nil, errs.NewUnexpectedError()
		}
	}

	m := repository.Member{
		CompanyId: req.CompanyId, Email: req.Email, Name: name, Role: role,
		Status: "active", UserId: &userId,
	}
	if _, err = s.memberRepository.Create(tx, m); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &CreateMemberResponse{
		Status: true, Desc: "Member added successfully",
		Data: CreateMemberData{TempPassword: tempPassword},
	}, nil
}

func (s memberService) Delete(id, userId int) (*SimpleResponse, error) {
	target, err := s.memberRepository.GetById(id)
	if err != nil {
		return nil, errs.NewNotFoundError("member not found")
	}

	requesterRole, err := permission.EffectiveRole(s.db, target.CompanyId, userId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	if !canRemove(requesterRole, target.Role) {
		return nil, errs.NewForbiddenError("access denied")
	}

	tx, err := s.memberRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.memberRepository.Delete(tx, id); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Member removed successfully"}, nil
}
