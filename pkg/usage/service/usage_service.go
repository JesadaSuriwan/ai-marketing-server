package service

import (
	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/pkg/usage/repository"
	"github.com/ai-marketing/ai-marketing-server/providers/usage"
)

type usageService struct {
	usageRepository repository.UsageRepository
}

func NewUsageService(usageRepository repository.UsageRepository) UsageService {
	return usageService{usageRepository}
}

func (s usageService) Log(companyId *int, engine, purpose, model string, u usage.Usage) error {
	return s.usageRepository.Log(repository.UsageLog{
		CompanyId: companyId, Engine: engine, Purpose: purpose, Model: model,
		InputTokens: u.InputTokens, OutputTokens: u.OutputTokens,
		CostUsd: usage.EstimateCost(model, u),
	})
}

func (s usageService) GetSummaryForCompany(companyId int, from, to string) (*UsageSummaryResponse, error) {
	rows, err := s.usageRepository.GetBreakdownForCompany(companyId, from, to)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	data := []UsageBreakdownData{}
	for _, r := range rows {
		data = append(data, UsageBreakdownData{
			Engine: r.Engine, Purpose: r.Purpose, Calls: r.Calls,
			InputTokens: r.InputTokens, OutputTokens: r.OutputTokens, CostUsd: r.CostUsd,
		})
	}

	return &UsageSummaryResponse{Status: true, Desc: "Get usage summary successful", Data: data}, nil
}
