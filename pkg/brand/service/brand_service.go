package service

import (
	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/pkg/brand/repository"
)

type brandService struct {
	brandRepository repository.BrandRepository
}

func NewBrandService(brandRepository repository.BrandRepository) BrandService {
	return brandService{brandRepository}
}

func toBrandData(b repository.Brand) BrandData {
	return BrandData{
		Id:          b.Id,
		CompanyId:   b.CompanyId,
		Name:        b.Name,
		Domain:      b.Domain,
		Industry:    b.Industry,
		Description: b.Description,
		Status:      b.Status,
		IsOwn:       b.IsOwn,
		LogoUrl:     b.LogoUrl,
		CreatedAt:   b.CreatedAt,
	}
}

func (s brandService) verifyOwnership(id, userId int) error {
	owned, err := s.brandRepository.BelongsToUser(id, userId)
	if err != nil || !owned {
		return errs.NewForbiddenError("access denied")
	}
	return nil
}

func (s brandService) GetAll(companyId int) (*BrandListResponse, error) {
	brands, err := s.brandRepository.GetAll(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	data := []BrandData{}
	for _, b := range brands {
		data = append(data, toBrandData(b))
	}

	return &BrandListResponse{Status: true, Desc: "Get brands successful", Data: data}, nil
}

func (s brandService) Create(req CreateBrandRequest) (*BrandResponse, error) {
	tx, err := s.brandRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	status := req.Status
	if status == "" {
		status = "active"
	}

	b := repository.Brand{
		CompanyId:   req.CompanyId,
		Name:        req.Name,
		Domain:      req.Domain,
		Industry:    req.Industry,
		Description: req.Description,
		Status:      status,
		IsOwn:       req.IsOwn,
		LogoUrl:     req.LogoUrl,
	}

	id, err := s.brandRepository.Create(tx, b)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if req.IsOwn {
		if err = s.brandRepository.ClearOwn(tx, req.CompanyId, id); err != nil {
			logs.Error(err)
			return nil, errs.NewUnexpectedError()
		}
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	b.Id = id
	return &BrandResponse{Status: true, Desc: "Brand created successfully", Data: toBrandData(b)}, nil
}

func (s brandService) GetById(id, userId int) (*BrandResponse, error) {
	if err := s.verifyOwnership(id, userId); err != nil {
		return nil, err
	}
	b, err := s.brandRepository.GetById(id)
	if err != nil {
		return nil, errs.NewNotFoundError("brand not found")
	}
	return &BrandResponse{Status: true, Desc: "Get brand successful", Data: toBrandData(*b)}, nil
}

func (s brandService) Update(id, userId int, req UpdateBrandRequest) (*SimpleResponse, error) {
	if err := s.verifyOwnership(id, userId); err != nil {
		return nil, err
	}

	existing, err := s.brandRepository.GetById(id)
	if err != nil {
		return nil, errs.NewNotFoundError("brand not found")
	}

	tx, err := s.brandRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	b := repository.Brand{
		Id:          id,
		Name:        req.Name,
		Domain:      req.Domain,
		Industry:    req.Industry,
		Description: req.Description,
		Status:      req.Status,
		IsOwn:       req.IsOwn,
		LogoUrl:     req.LogoUrl,
	}

	if err = s.brandRepository.Update(tx, b); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if req.IsOwn {
		if err = s.brandRepository.ClearOwn(tx, existing.CompanyId, id); err != nil {
			logs.Error(err)
			return nil, errs.NewUnexpectedError()
		}
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Brand updated successfully"}, nil
}

func (s brandService) Delete(id, userId int) (*SimpleResponse, error) {
	if err := s.verifyOwnership(id, userId); err != nil {
		return nil, err
	}

	tx, err := s.brandRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.brandRepository.Delete(tx, id); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Brand deleted successfully"}, nil
}

func (s brandService) UpdateStatus(id, userId int, status string) (*SimpleResponse, error) {
	if err := s.verifyOwnership(id, userId); err != nil {
		return nil, err
	}

	tx, err := s.brandRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.brandRepository.UpdateStatus(tx, id, status); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Brand status updated successfully"}, nil
}
