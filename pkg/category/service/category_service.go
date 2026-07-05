package service

import (
	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/pkg/category/repository"
)

type categoryService struct {
	categoryRepository repository.CategoryRepository
}

func NewCategoryService(categoryRepository repository.CategoryRepository) CategoryService {
	return categoryService{categoryRepository}
}

func (s categoryService) verifyOwnership(id, userId int) error {
	owned, err := s.categoryRepository.BelongsToUser(id, userId)
	if err != nil || !owned {
		return errs.NewForbiddenError("access denied")
	}
	return nil
}

func (s categoryService) GetAll(companyId int) (*CategoryListResponse, error) {
	categories, err := s.categoryRepository.GetAll(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	data := []CategoryData{}
	for _, c := range categories {
		data = append(data, CategoryData{Id: c.Id, CompanyId: c.CompanyId, Name: c.Name, CreatedAt: c.CreatedAt})
	}

	return &CategoryListResponse{Status: true, Desc: "Get categories successful", Data: data}, nil
}

func (s categoryService) Create(req CreateCategoryRequest) (*SimpleResponse, error) {
	tx, err := s.categoryRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	c := repository.Category{CompanyId: req.CompanyId, Name: req.Name}
	if _, err = s.categoryRepository.Create(tx, c); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Category created successfully"}, nil
}

func (s categoryService) Update(id, userId int, req UpdateCategoryRequest) (*SimpleResponse, error) {
	if err := s.verifyOwnership(id, userId); err != nil {
		return nil, err
	}

	tx, err := s.categoryRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.categoryRepository.Update(tx, id, req.Name); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Category updated successfully"}, nil
}

func (s categoryService) Delete(id, userId int) (*SimpleResponse, error) {
	if err := s.verifyOwnership(id, userId); err != nil {
		return nil, err
	}

	tx, err := s.categoryRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.categoryRepository.Delete(tx, id); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Category deleted successfully"}, nil
}
