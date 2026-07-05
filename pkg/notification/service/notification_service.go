package service

import (
	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/pkg/notification/repository"
)

type notificationService struct {
	notificationRepository repository.NotificationRepository
}

func NewNotificationService(notificationRepository repository.NotificationRepository) NotificationService {
	return notificationService{notificationRepository}
}

func (s notificationService) verifyOwnership(id, userId int) error {
	ok, err := s.notificationRepository.BelongsToUser(id, userId)
	if err != nil {
		logs.Error(err)
		return errs.NewUnexpectedError()
	}
	if !ok {
		return errs.NewForbiddenError("access denied")
	}
	return nil
}

func (s notificationService) GetAll(companyId int) (*NotificationListResponse, error) {
	notifications, err := s.notificationRepository.GetAll(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	data := []NotificationData{}
	for _, n := range notifications {
		data = append(data, NotificationData{
			Id: n.Id, UserId: n.UserId, CompanyId: n.CompanyId, Type: n.Type,
			Title: n.Title, Description: n.Description, IsRead: n.IsRead, CreatedAt: n.CreatedAt,
		})
	}

	return &NotificationListResponse{Status: true, Desc: "Get notifications successful", Data: data}, nil
}

func (s notificationService) Create(req CreateNotificationRequest) (*SimpleResponse, error) {
	tx, err := s.notificationRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	n := repository.Notification{
		UserId: req.UserId, CompanyId: req.CompanyId, Type: req.Type,
		Title: req.Title, Description: req.Description,
	}

	if _, err = s.notificationRepository.Create(tx, n); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Notification created successfully"}, nil
}

func (s notificationService) MarkRead(id, userId int) (*SimpleResponse, error) {
	if err := s.verifyOwnership(id, userId); err != nil {
		return nil, err
	}

	tx, err := s.notificationRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.notificationRepository.MarkRead(tx, id); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Notification marked as read"}, nil
}

func (s notificationService) MarkAllRead(companyId int) (*SimpleResponse, error) {
	tx, err := s.notificationRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.notificationRepository.MarkAllRead(tx, companyId); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "All notifications marked as read"}, nil
}

func (s notificationService) Delete(id, userId int) (*SimpleResponse, error) {
	if err := s.verifyOwnership(id, userId); err != nil {
		return nil, err
	}

	tx, err := s.notificationRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.notificationRepository.Delete(tx, id); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Notification deleted successfully"}, nil
}
