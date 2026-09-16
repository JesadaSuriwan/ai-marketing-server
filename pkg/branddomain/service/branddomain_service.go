package service

import (
	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/pkg/branddomain/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/permission"
	"github.com/jmoiron/sqlx"
)

type brandDomainService struct {
	brandDomainRepository repository.BrandDomainRepository
	db                    *sqlx.DB
}

func NewBrandDomainService(brandDomainRepository repository.BrandDomainRepository, db *sqlx.DB) BrandDomainService {
	return brandDomainService{brandDomainRepository, db}
}

// requireEditRole is shared by Create/Delete — Admin, Team Lead, and
// Specialist can manage competitor domains; Customer cannot.
func (s brandDomainService) requireEditRole(companyId, userId int) error {
	role, err := permission.EffectiveRole(s.db, companyId, userId)
	if err != nil {
		logs.Error(err)
		return errs.NewUnexpectedError()
	}
	if !permission.Allowed(role, permission.Admin, permission.TeamLead, permission.Specialist) {
		return errs.NewForbiddenError("access denied")
	}
	return nil
}

func (s brandDomainService) GetAll(brandId int) (*BrandDomainListResponse, error) {
	domains, err := s.brandDomainRepository.GetAll(brandId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	data := []BrandDomainData{}
	for _, d := range domains {
		data = append(data, BrandDomainData{Id: d.Id, BrandId: d.BrandId, Domain: d.Domain, CreatedAt: d.CreatedAt})
	}

	return &BrandDomainListResponse{Status: true, Desc: "Get brand domains successful", Data: data}, nil
}

func (s brandDomainService) Create(userId int, req CreateBrandDomainRequest) (*SimpleResponse, error) {
	companyId, err := s.brandDomainRepository.GetCompanyIdByBrandId(req.BrandId)
	if err != nil {
		return nil, errs.NewNotFoundError("brand not found")
	}
	if err := s.requireEditRole(companyId, userId); err != nil {
		return nil, err
	}

	tx, err := s.brandDomainRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	d := repository.BrandDomain{BrandId: req.BrandId, Domain: req.Domain}
	if _, err = s.brandDomainRepository.Create(tx, d); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Domain added successfully"}, nil
}

func (s brandDomainService) Delete(id, userId int) (*SimpleResponse, error) {
	d, err := s.brandDomainRepository.GetById(id)
	if err != nil {
		return nil, errs.NewNotFoundError("domain not found")
	}
	companyId, err := s.brandDomainRepository.GetCompanyIdByBrandId(d.BrandId)
	if err != nil {
		return nil, errs.NewNotFoundError("brand not found")
	}
	if err := s.requireEditRole(companyId, userId); err != nil {
		return nil, err
	}

	tx, err := s.brandDomainRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.brandDomainRepository.Delete(tx, id); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Domain removed successfully"}, nil
}
