package service

import (
	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	brandRepository "github.com/ai-marketing/ai-marketing-server/pkg/brand/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/brandcandidate/repository"
)

type brandCandidateService struct {
	brandCandidateRepository repository.BrandCandidateRepository
	brandRepository          brandRepository.BrandRepository
}

func NewBrandCandidateService(
	brandCandidateRepository repository.BrandCandidateRepository,
	brandRepository brandRepository.BrandRepository,
) BrandCandidateService {
	return brandCandidateService{brandCandidateRepository, brandRepository}
}

func toData(c repository.BrandCandidate) BrandCandidateData {
	return BrandCandidateData{
		Id: c.Id, CompanyId: c.CompanyId, Name: c.Name, MentionCount: c.MentionCount,
		FirstSeen: c.FirstSeen, LastSeen: c.LastSeen,
	}
}

func (s brandCandidateService) List(companyId int) (*BrandCandidateListResponse, error) {
	rows, err := s.brandCandidateRepository.GetPending(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	data := []BrandCandidateData{}
	for _, r := range rows {
		data = append(data, toData(r))
	}
	return &BrandCandidateListResponse{Status: true, Desc: "Get brand candidates successful", Data: data}, nil
}

func (s brandCandidateService) Resolve(candidateId int, req ResolveRequest) (*SimpleResponse, error) {
	candidate, err := s.brandCandidateRepository.GetById(candidateId)
	if err != nil {
		return nil, errs.NewNotFoundError("brand candidate not found")
	}
	if candidate.CompanyId != req.CompanyId {
		return nil, errs.NewForbiddenError("access denied")
	}

	tx, err := s.brandCandidateRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	switch req.Action {
	case "own_alias":
		brands, err := s.brandRepository.GetAll(req.CompanyId)
		if err != nil {
			logs.Error(err)
			return nil, errs.NewUnexpectedError()
		}
		var ownId int
		for _, b := range brands {
			if b.IsOwn {
				ownId = b.Id
				break
			}
		}
		if ownId == 0 {
			return nil, errs.NewBadRequestError("no own brand set for this company yet — add one in Settings first")
		}
		if err := s.brandRepository.AddAlias(tx, ownId, candidate.Name); err != nil {
			logs.Error(err)
			return nil, errs.NewUnexpectedError()
		}

	case "merge_competitor":
		if req.TargetBrandId == nil {
			return nil, errs.NewBadRequestError("target_brand_id is required for merge_competitor")
		}
		target, err := s.brandRepository.GetById(*req.TargetBrandId)
		if err != nil || target.CompanyId != req.CompanyId || target.IsOwn {
			return nil, errs.NewBadRequestError("target_brand_id must be an existing competitor of this company")
		}
		if err := s.brandRepository.AddAlias(tx, target.Id, candidate.Name); err != nil {
			logs.Error(err)
			return nil, errs.NewUnexpectedError()
		}

	case "new_competitor":
		if _, err := s.brandRepository.Create(tx, brandRepository.Brand{
			CompanyId: req.CompanyId, Name: candidate.Name, Domain: "", Status: "active", IsOwn: false,
		}); err != nil {
			logs.Error(err)
			return nil, errs.NewUnexpectedError()
		}

	case "dismiss":
		// no brand/alias side effect — just stop resurfacing it

	default:
		return nil, errs.NewBadRequestError("unknown action")
	}

	status := "resolved"
	if req.Action == "dismiss" {
		status = "dismissed"
	}
	if err := s.brandCandidateRepository.SetStatus(tx, candidateId, status); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err := tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Brand candidate resolved successfully"}, nil
}
