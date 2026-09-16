package service

import (
	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/pkg/recommendationrule/repository"
)

type recommendationRuleService struct {
	recommendationRuleRepository repository.RecommendationRuleRepository
}

func NewRecommendationRuleService(recommendationRuleRepository repository.RecommendationRuleRepository) RecommendationRuleService {
	return recommendationRuleService{recommendationRuleRepository}
}

func toData(r repository.RecommendationRule) RecommendationRuleData {
	return RecommendationRuleData{
		Id: r.Id, CompanyId: r.CompanyId, Text: r.Text, Active: r.Active,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func (s recommendationRuleService) verifyOwnership(id, userId int) error {
	owned, err := s.recommendationRuleRepository.BelongsToUser(id, userId)
	if err != nil || !owned {
		return errs.NewForbiddenError("access denied")
	}
	return nil
}

func (s recommendationRuleService) GetAll(companyId int) (*RecommendationRuleListResponse, error) {
	rules, err := s.recommendationRuleRepository.GetAll(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	data := []RecommendationRuleData{}
	for _, r := range rules {
		data = append(data, toData(r))
	}

	return &RecommendationRuleListResponse{Status: true, Desc: "Get recommendation rules successful", Data: data}, nil
}

func (s recommendationRuleService) Create(req CreateRecommendationRuleRequest) (*SimpleResponse, error) {
	tx, err := s.recommendationRuleRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	r := repository.RecommendationRule{CompanyId: req.CompanyId, Text: req.Text}
	if _, err = s.recommendationRuleRepository.Create(tx, r); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Rule created successfully"}, nil
}

func (s recommendationRuleService) Update(id, userId int, req UpdateRecommendationRuleRequest) (*SimpleResponse, error) {
	if err := s.verifyOwnership(id, userId); err != nil {
		return nil, err
	}

	tx, err := s.recommendationRuleRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.recommendationRuleRepository.Update(tx, id, req.Text, req.Active); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Rule updated successfully"}, nil
}

func (s recommendationRuleService) Delete(id, userId int) (*SimpleResponse, error) {
	if err := s.verifyOwnership(id, userId); err != nil {
		return nil, err
	}

	tx, err := s.recommendationRuleRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.recommendationRuleRepository.Delete(tx, id); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Rule deleted successfully"}, nil
}
