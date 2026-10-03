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
	// encryptionKey backs the admin-viewable password copy for Customer role
	// members only — see utils.EncryptPassword.
	encryptionKey string
}

func NewMemberService(
	memberRepository repository.MemberRepository,
	authRepository authRepository.AuthRepository,
	companyRepository companyRepository.CompanyRepository,
	db *sqlx.DB,
	encryptionKey string,
) MemberService {
	return memberService{memberRepository, authRepository, companyRepository, db, encryptionKey}
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

	// Normalized here (not just inside authRepository.GetUserByEmail) because
	// this value also gets written straight to company_members.email below,
	// which isn't covered by that repository's own normalization.
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

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

		// Customer accounts: the Admin/Team Lead's own chosen password (if
		// given) is what the client actually uses to log in — no forced
		// change, since this isn't a colleague picking their own password,
		// it's a credential handed to an external client. Every other role
		// keeps the existing random-temp-password-then-forced-change flow.
		isCustomer := role == permission.Customer
		mustChangePassword := true
		var encryptedPassword *string

		if isCustomer {
			mustChangePassword = false
			if strings.TrimSpace(req.Password) != "" {
				tempPassword = req.Password
			} else {
				tempPassword, err = utils.GenerateTempPassword()
				if err != nil {
					logs.Error(err)
					return nil, errs.NewUnexpectedError()
				}
			}
			if encrypted, encErr := utils.EncryptPassword(tempPassword, s.encryptionKey); encErr == nil {
				encryptedPassword = &encrypted
			} else {
				// Don't fail account creation over the view/edit convenience
				// feature being unconfigured — the account still works with
				// the password just set, it just won't be viewable later
				// until MEMBER_PASSWORD_ENCRYPTION_KEY is set and it's reset.
				logs.Error(encErr)
			}
		} else {
			tempPassword, err = utils.GenerateTempPassword()
			if err != nil {
				logs.Error(err)
				return nil, errs.NewUnexpectedError()
			}
		}

		hashed, err := utils.HashPassword(tempPassword)
		if err != nil {
			logs.Error(err)
			return nil, errs.NewUnexpectedError()
		}

		userId, err = s.authRepository.CreateUser(tx, authRepository.User{
			Email: req.Email, Password: hashed, Name: name, Initials: initials,
			MustChangePassword: mustChangePassword, EncryptedPassword: encryptedPassword,
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

// requireCustomerTargetManager loads the target member and checks both that
// the requester is allowed to manage it (Admin/Team Lead only — the same
// gate as Create, deliberately not extended to Specialist even though
// Specialist can remove a Customer, since setting/viewing a password is more
// consequential) and that the target is actually a Customer — password
// view/edit is scoped to exactly that role, not Team Lead/Specialist/Admin
// accounts, which manage their own passwords.
func (s memberService) requireCustomerTargetManager(id, requesterUserId int) (*repository.Member, error) {
	target, err := s.memberRepository.GetById(id)
	if err != nil {
		return nil, errs.NewNotFoundError("member not found")
	}
	requesterRole, err := permission.EffectiveRole(s.db, target.CompanyId, requesterUserId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	if !permission.Allowed(requesterRole, permission.Admin, permission.TeamLead) {
		return nil, errs.NewForbiddenError("access denied")
	}
	if target.Role != permission.Customer {
		return nil, errs.NewBadRequestError("password management is only available for Customer role members")
	}
	if target.UserId == nil {
		return nil, errs.NewBadRequestError("this member hasn't accepted their invite yet")
	}
	return target, nil
}

func (s memberService) SetPassword(id, requesterUserId int, req SetMemberPasswordRequest) (*SimpleResponse, error) {
	target, err := s.requireCustomerTargetManager(id, requesterUserId)
	if err != nil {
		return nil, err
	}

	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	encrypted, err := utils.EncryptPassword(req.Password, s.encryptionKey)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewBadRequestError("password encryption is not configured on the server")
	}

	tx, err := s.authRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.authRepository.UpdatePassword(tx, *target.UserId, hashed); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	if err = s.authRepository.UpdateEncryptedPassword(tx, *target.UserId, encrypted); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Password updated successfully"}, nil
}

func (s memberService) GetPassword(id, requesterUserId int) (*MemberPasswordResponse, error) {
	target, err := s.requireCustomerTargetManager(id, requesterUserId)
	if err != nil {
		return nil, err
	}

	user, err := s.authRepository.GetUserById(*target.UserId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	if user.EncryptedPassword == nil {
		return &MemberPasswordResponse{Status: true, Desc: "No viewable password set", Data: MemberPasswordData{Password: nil}}, nil
	}

	decrypted, err := utils.DecryptPassword(*user.EncryptedPassword, s.encryptionKey)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &MemberPasswordResponse{Status: true, Desc: "Get password successful", Data: MemberPasswordData{Password: &decrypted}}, nil
}
