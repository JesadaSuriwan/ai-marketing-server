package service

import (
	"strings"

	"github.com/ai-marketing/ai-marketing-server/constants"
	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/pkg/auth/repository"
	"github.com/ai-marketing/ai-marketing-server/utils"
)

type authService struct {
	authRepository repository.AuthRepository
	// encryptionKey backs re-syncing a Customer role member's admin-viewable
	// password copy whenever they change their own password — see
	// ChangePassword below and utils.EncryptPassword for why this exists.
	encryptionKey string
}

func NewAuthService(authRepository repository.AuthRepository, encryptionKey string) AuthService {
	return authService{authRepository, encryptionKey}
}

func (s authService) Register(req RegisterRequest) (*AuthResponse, string, error) {
	// authRepository normalizes internally too, but doing it here as well
	// keeps req.Email (used below in the token and response) consistent with
	// what's actually stored, rather than echoing back whatever casing the
	// caller happened to type.
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	existing, _ := s.authRepository.GetUserByEmail(req.Email)
	if existing != nil {
		return nil, "", errs.NewBadRequestError("email already registered")
	}

	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		logs.Error(err)
		return nil, "", errs.NewUnexpectedError()
	}

	tx, err := s.authRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, "", errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	newUser := repository.User{
		Email:    req.Email,
		Password: hashed,
		Name:     req.Name,
		Initials: req.Initials,
	}

	id, err := s.authRepository.CreateUser(tx, newUser)
	if err != nil {
		logs.Error(err)
		return nil, "", errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, "", errs.NewUnexpectedError()
	}

	token, err := utils.GenerateToken(req.Email, id, constants.RoleUser)
	if err != nil {
		logs.Error(err)
		return nil, "", errs.NewUnexpectedError()
	}

	response := &AuthResponse{
		Status: true,
		Desc:   "Register successful",
		Data: UserData{
			Id:                 id,
			Email:              req.Email,
			Name:               req.Name,
			Initials:           req.Initials,
			MustChangePassword: false,
		},
	}

	return response, token, nil
}

func (s authService) Login(req LoginRequest) (*AuthResponse, string, error) {
	user, err := s.authRepository.GetUserByEmail(req.Email)
	if err != nil {
		return nil, "", errs.NewBadRequestError("invalid email or password")
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, "", errs.NewBadRequestError("invalid email or password")
	}

	token, err := utils.GenerateToken(user.Email, user.Id, constants.RoleUser)
	if err != nil {
		logs.Error(err)
		return nil, "", errs.NewUnexpectedError()
	}

	response := &AuthResponse{
		Status: true,
		Desc:   "Login successful",
		Data: UserData{
			Id:                 user.Id,
			Email:              user.Email,
			Name:               user.Name,
			Initials:           user.Initials,
			MustChangePassword: user.MustChangePassword,
		},
	}

	return response, token, nil
}

func (s authService) Me(userId int) (*AuthResponse, error) {
	user, err := s.authRepository.GetUserById(userId)
	if err != nil {
		return nil, errs.NewNotFoundError("user not found")
	}

	response := &AuthResponse{
		Status: true,
		Desc:   "Get user successful",
		Data: UserData{
			Id:                 user.Id,
			Email:              user.Email,
			Name:               user.Name,
			Initials:           user.Initials,
			MustChangePassword: user.MustChangePassword,
		},
	}

	return response, nil
}

// ChangePassword always requires the current password — this covers both a
// voluntary change and the forced first change after logging in with a
// temp password (the temp password doubles as "current" in that case), so
// there's a single consistent security model instead of two code paths.
func (s authService) ChangePassword(userId int, req ChangePasswordRequest) (*SimpleResponse, error) {
	user, err := s.authRepository.GetUserById(userId)
	if err != nil {
		return nil, errs.NewNotFoundError("user not found")
	}

	if !utils.CheckPassword(req.CurrentPassword, user.Password) {
		return nil, errs.NewBadRequestError("current password is incorrect")
	}

	hashed, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	tx, err := s.authRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.authRepository.UpdatePassword(tx, userId, hashed); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	// Only re-sync the admin-viewable copy if this account already had one —
	// i.e. only Customer role members, who are the only accounts this ever
	// gets set for in the first place. A regular team member calling this
	// endpoint never starts having their password encrypted/stored; nil
	// stays nil.
	if user.EncryptedPassword != nil {
		if encrypted, encErr := utils.EncryptPassword(req.NewPassword, s.encryptionKey); encErr == nil {
			if err = s.authRepository.UpdateEncryptedPassword(tx, userId, encrypted); err != nil {
				logs.Error(err)
				return nil, errs.NewUnexpectedError()
			}
		} else {
			// Key not configured or otherwise broken — don't block the user
			// from changing their own password over a feature they don't
			// know exists; just log it so it's visible to us.
			logs.Error(encErr)
		}
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Password changed successfully"}, nil
}
