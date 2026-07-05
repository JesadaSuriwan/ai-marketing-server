package service

import (
	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/pkg/prompt/repository"
)

type promptService struct {
	promptRepository repository.PromptRepository
}

func NewPromptService(promptRepository repository.PromptRepository) PromptService {
	return promptService{promptRepository}
}

func toPromptData(p repository.Prompt) PromptData {
	return PromptData{Id: p.Id, CompanyId: p.CompanyId, CategoryId: p.CategoryId, Title: p.Title, Content: p.Content, CreatedAt: p.CreatedAt}
}

func (s promptService) verifyOwnership(id, userId int) error {
	owned, err := s.promptRepository.BelongsToUser(id, userId)
	if err != nil || !owned {
		return errs.NewForbiddenError("access denied")
	}
	return nil
}

func (s promptService) GetAll(companyId int) (*PromptListResponse, error) {
	prompts, err := s.promptRepository.GetAll(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	data := []PromptData{}
	for _, p := range prompts {
		data = append(data, toPromptData(p))
	}

	return &PromptListResponse{Status: true, Desc: "Get prompts successful", Data: data}, nil
}

func (s promptService) Create(req CreatePromptRequest) (*PromptResponse, error) {
	tx, err := s.promptRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	p := repository.Prompt{
		CompanyId:  req.CompanyId,
		CategoryId: req.CategoryId,
		Title:      req.Title,
		Content:    req.Content,
	}

	id, err := s.promptRepository.Create(tx, p)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	p.Id = id
	return &PromptResponse{Status: true, Desc: "Prompt created successfully", Data: toPromptData(p)}, nil
}

func (s promptService) GetById(id, userId int) (*PromptResponse, error) {
	if err := s.verifyOwnership(id, userId); err != nil {
		return nil, err
	}
	p, err := s.promptRepository.GetById(id)
	if err != nil {
		return nil, errs.NewNotFoundError("prompt not found")
	}
	return &PromptResponse{Status: true, Desc: "Get prompt successful", Data: toPromptData(*p)}, nil
}

func (s promptService) Update(id, userId int, req UpdatePromptRequest) (*SimpleResponse, error) {
	if err := s.verifyOwnership(id, userId); err != nil {
		return nil, err
	}

	tx, err := s.promptRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	p := repository.Prompt{Id: id, CategoryId: req.CategoryId, Title: req.Title, Content: req.Content}
	if err = s.promptRepository.Update(tx, p); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Prompt updated successfully"}, nil
}

func (s promptService) Delete(id, userId int) (*SimpleResponse, error) {
	if err := s.verifyOwnership(id, userId); err != nil {
		return nil, err
	}

	tx, err := s.promptRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.promptRepository.Delete(tx, id); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Prompt deleted successfully"}, nil
}
