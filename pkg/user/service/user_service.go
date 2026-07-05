package service

import (
	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/pkg/user/repository"
)

type userService struct {
	userRepository repository.UserRepository
}

func NewUserService(userRepository repository.UserRepository) UserService {
	return userService{userRepository}
}

func (s userService) GetById(id int) (*UserResponse, error) {
	user, err := s.userRepository.GetById(id)
	if err != nil {
		return nil, errs.NewNotFoundError("user not found")
	}

	return &UserResponse{
		Status: true,
		Desc:   "Get user successful",
		Data: UserData{
			Id:       user.Id,
			Email:    user.Email,
			Name:     user.Name,
			Initials: user.Initials,
		},
	}, nil
}

func (s userService) Update(id int, req UpdateUserRequest) (*SimpleResponse, error) {
	tx, err := s.userRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.userRepository.Update(tx, id, req.Name, req.Initials); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "User updated successfully"}, nil
}
