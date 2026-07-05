package service

import (
	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/pkg/subdomain/repository"
)

type subdomainService struct {
	subdomainRepository repository.SubdomainRepository
}

func NewSubdomainService(subdomainRepository repository.SubdomainRepository) SubdomainService {
	return subdomainService{subdomainRepository}
}

func (s subdomainService) verifyOwnership(id, userId int) error {
	ok, err := s.subdomainRepository.BelongsToUser(id, userId)
	if err != nil {
		logs.Error(err)
		return errs.NewUnexpectedError()
	}
	if !ok {
		return errs.NewForbiddenError("access denied")
	}
	return nil
}

func (s subdomainService) GetAll(companyId int) (*SubdomainListResponse, error) {
	subdomains, err := s.subdomainRepository.GetAll(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	data := []SubdomainData{}
	for _, sub := range subdomains {
		data = append(data, SubdomainData{Id: sub.Id, CompanyId: sub.CompanyId, Subdomain: sub.Subdomain, Status: sub.Status, CreatedAt: sub.CreatedAt})
	}

	return &SubdomainListResponse{Status: true, Desc: "Get subdomains successful", Data: data}, nil
}

func (s subdomainService) Create(req CreateSubdomainRequest) (*SimpleResponse, error) {
	tx, err := s.subdomainRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	sub := repository.Subdomain{CompanyId: req.CompanyId, Subdomain: req.Subdomain, Status: "Active"}
	if _, err = s.subdomainRepository.Create(tx, sub); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Subdomain created successfully"}, nil
}

func (s subdomainService) Delete(id, userId int) (*SimpleResponse, error) {
	if err := s.verifyOwnership(id, userId); err != nil {
		return nil, err
	}

	tx, err := s.subdomainRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.subdomainRepository.Delete(tx, id); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Subdomain deleted successfully"}, nil
}
