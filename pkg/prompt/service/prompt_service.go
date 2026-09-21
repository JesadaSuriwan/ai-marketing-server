package service

import (
	"fmt"
	"strings"

	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	companyRepository "github.com/ai-marketing/ai-marketing-server/pkg/company/repository"
	"github.com/ai-marketing/ai-marketing-server/pkg/permission"
	"github.com/ai-marketing/ai-marketing-server/pkg/prompt/repository"
	promptRunService "github.com/ai-marketing/ai-marketing-server/pkg/promptrun/service"
	"github.com/jmoiron/sqlx"
)

type promptService struct {
	promptRepository  repository.PromptRepository
	companyRepository companyRepository.CompanyRepository
	promptRunService  promptRunService.PromptRunService
	db                *sqlx.DB
}

func NewPromptService(
	promptRepository repository.PromptRepository,
	companyRepository companyRepository.CompanyRepository,
	promptRunService promptRunService.PromptRunService,
	db *sqlx.DB,
) PromptService {
	return promptService{promptRepository, companyRepository, promptRunService, db}
}

func toPromptData(p repository.Prompt) PromptData {
	countries := []string{}
	if p.Countries != "" {
		countries = strings.Split(p.Countries, ",")
	}
	return PromptData{
		Id: p.Id, CompanyId: p.CompanyId, TagId: p.TagId, TagName: p.TagName, Countries: countries,
		Title: p.Title, Content: p.Content, Active: p.Active, CreatedAt: p.CreatedAt,
	}
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

// Each prompt runs for exactly one country: the engines are told where the
// user is, so a multi-country prompt would blend markets. Track another
// country by adding the prompt again.
func validateSingleCountry(codes []string) error {
	if len(codes) != 1 || len(strings.TrimSpace(codes[0])) != 2 {
		return errs.NewBadRequestError("choose exactly one country for a prompt")
	}
	return nil
}

func (s promptService) Create(req CreatePromptRequest) (*PromptResponse, error) {
	if err := validateSingleCountry(req.CountryCodes); err != nil {
		return nil, err
	}
	company, err := s.companyRepository.GetById(req.CompanyId)
	if err != nil {
		return nil, errs.NewNotFoundError("company not found")
	}
	activeCount, err := s.promptRepository.CountActive(req.CompanyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	if activeCount >= company.PromptLimit {
		return nil, errs.NewBadRequestError(fmt.Sprintf(
			"you've reached your active prompt limit of %d — deactivate a prompt or raise the limit in Settings",
			company.PromptLimit,
		))
	}

	tx, err := s.promptRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	p := repository.Prompt{
		CompanyId: req.CompanyId,
		TagId:     req.TagId,
		Title:     req.Title,
		Content:   req.Content,
	}

	id, err := s.promptRepository.Create(tx, p)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = s.promptRepository.SetCountries(tx, id, req.CountryCodes); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	// Best-effort, fire-and-forget: give the new prompt real data right away
	// instead of waiting for the next scheduled sweep. A full multi-engine
	// run can take 10-30s, so this must not block the HTTP response.
	go func() {
		if _, err := s.promptRunService.RunSystem(id); err != nil {
			logs.Error(fmt.Errorf("auto first-run failed for prompt %d: %w", id, err))
		}
	}()

	created, err := s.promptRepository.GetById(id)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	return &PromptResponse{Status: true, Desc: "Prompt created successfully", Data: toPromptData(*created)}, nil
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
	if req.CountryCodes != nil {
		if err := validateSingleCountry(req.CountryCodes); err != nil {
			return nil, err
		}
	}

	tx, err := s.promptRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.promptRepository.Update(tx, id, req.TagId, req.Title, req.Content); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if req.CountryCodes != nil {
		if err = s.promptRepository.SetCountries(tx, id, req.CountryCodes); err != nil {
			logs.Error(err)
			return nil, errs.NewUnexpectedError()
		}
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Prompt updated successfully"}, nil
}

// SetActive is gated to Specialist-and-above — Customers are view-only and
// shouldn't be able to pause/resume a prompt even via a direct API call.
func (s promptService) SetActive(id, userId int, active bool) (*SimpleResponse, error) {
	p, err := s.promptRepository.GetById(id)
	if err != nil {
		return nil, errs.NewNotFoundError("prompt not found")
	}

	role, err := permission.EffectiveRole(s.db, p.CompanyId, userId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	if !permission.Allowed(role, permission.Admin, permission.TeamLead, permission.Specialist) {
		return nil, errs.NewForbiddenError("access denied")
	}

	// Reactivating counts against the limit too — otherwise the cap could be
	// bypassed by creating prompts, pausing some, then resuming them all.
	if active && !p.Active {
		company, err := s.companyRepository.GetById(p.CompanyId)
		if err != nil {
			return nil, errs.NewNotFoundError("company not found")
		}
		activeCount, err := s.promptRepository.CountActive(p.CompanyId)
		if err != nil {
			logs.Error(err)
			return nil, errs.NewUnexpectedError()
		}
		if activeCount >= company.PromptLimit {
			return nil, errs.NewBadRequestError(fmt.Sprintf(
				"you've reached your active prompt limit of %d — deactivate another prompt or raise the limit in Settings",
				company.PromptLimit,
			))
		}
	}

	tx, err := s.promptRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.promptRepository.SetActive(tx, id, active); err != nil {
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
