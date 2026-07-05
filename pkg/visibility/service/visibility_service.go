package service

import (
	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/pkg/visibility/repository"
)

type visibilityService struct {
	visibilityRepository repository.VisibilityRepository
}

func NewVisibilityService(visibilityRepository repository.VisibilityRepository) VisibilityService {
	return visibilityService{visibilityRepository}
}

func (s visibilityService) GetAll(brandId int, from, to string) (*VisibilityListResponse, error) {
	items, err := s.visibilityRepository.GetAll(brandId, from, to)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	data := []VisibilityData{}
	for _, v := range items {
		data = append(data, VisibilityData{
			Id: v.Id, BrandId: v.BrandId, PromptId: v.PromptId, Platform: v.Platform,
			Score: v.Score, Mentions: v.Mentions, Date: v.Date, CreatedAt: v.CreatedAt,
		})
	}

	return &VisibilityListResponse{Status: true, Desc: "Get visibility successful", Data: data}, nil
}

func (s visibilityService) Create(req CreateVisibilityRequest) (*SimpleResponse, error) {
	tx, err := s.visibilityRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	v := repository.Visibility{
		BrandId:  req.BrandId,
		PromptId: req.PromptId,
		Platform: req.Platform,
		Score:    req.Score,
		Mentions: req.Mentions,
		Date:     req.Date,
	}

	if _, err = s.visibilityRepository.Create(tx, v); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Visibility created successfully"}, nil
}

func (s visibilityService) GetBreakdown(brandId int) (*BreakdownResponse, error) {
	items, err := s.visibilityRepository.GetBreakdown(brandId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	data := []PlatformBreakdownData{}
	for _, item := range items {
		data = append(data, PlatformBreakdownData{Platform: item.Platform, Score: item.Score, Mentions: item.Mentions})
	}

	return &BreakdownResponse{Status: true, Desc: "Get breakdown successful", Data: data}, nil
}
