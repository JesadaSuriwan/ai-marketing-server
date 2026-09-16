package service

import (
	"github.com/ai-marketing/ai-marketing-server/constants"
	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/pkg/auth/repository"
	"github.com/ai-marketing/ai-marketing-server/utils"
)

type authService struct {
	authRepository repository.AuthRepository
}

func NewAuthService(authRepository repository.AuthRepository) AuthService {
	return authService{authRepository}
}

func (s authService) Register(req RegisterRequest) (*AuthResponse, string, error) {
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

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Password changed successfully"}, nil
}
