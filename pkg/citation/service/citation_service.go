package service

import (
	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/pkg/citation/repository"
)

type citationService struct {
	citationRepository repository.CitationRepository
}

func NewCitationService(citationRepository repository.CitationRepository) CitationService {
	return citationService{citationRepository}
}

func toCitationData(c repository.Citation, competitors []repository.CitationCompetitor, history []repository.CitationRankingHistory) CitationData {
	competitorList := []CompetitorDTO{}
	for _, cc := range competitors {
		competitorList = append(competitorList, CompetitorDTO{Id: cc.Id, CompetitorName: cc.CompetitorName})
	}

	historyList := []RankingHistoryDTO{}
	for _, h := range history {
		historyList = append(historyList, RankingHistoryDTO{Id: h.Id, DateLabel: h.DateLabel, Rank: h.Rank})
	}

	return CitationData{
		Id: c.Id, PromptId: c.PromptId, BrandId: c.BrandId, Ranking: c.Ranking, Content: c.Content,
		Url: c.Url, AiPlatform: c.AiPlatform, DateDiscovered: c.DateDiscovered, LastChecked: c.LastChecked,
		Sentiment: c.Sentiment, BrandPositioning: c.BrandPositioning, CitationFrequency: c.CitationFrequency,
		Snippet: c.Snippet, IsCompetitor: c.IsCompetitor, Notes: c.Notes, IsArchived: c.IsArchived,
		Competitors: competitorList, RankingHistory: historyList, CreatedAt: c.CreatedAt,
	}
}

func (s citationService) verifyOwnership(id, userId int) error {
	owned, err := s.citationRepository.BelongsToUser(id, userId)
	if err != nil || !owned {
		return errs.NewForbiddenError("access denied")
	}
	return nil
}

func (s citationService) GetAll(promptId int) (*CitationListResponse, error) {
	citations, err := s.citationRepository.GetAll(promptId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	data := []CitationData{}
	for _, c := range citations {
		data = append(data, toCitationData(c, []repository.CitationCompetitor{}, []repository.CitationRankingHistory{}))
	}

	return &CitationListResponse{Status: true, Desc: "Get citations successful", Data: data}, nil
}

func (s citationService) Create(req CreateCitationRequest) (*SimpleResponse, error) {
	tx, err := s.citationRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	sentiment := req.Sentiment
	if sentiment == "" {
		sentiment = "neutral"
	}
	positioning := req.BrandPositioning
	if positioning == "" {
		positioning = "mentioned"
	}

	c := repository.Citation{
		PromptId: req.PromptId, BrandId: req.BrandId, Ranking: req.Ranking, Content: req.Content,
		Url: req.Url, AiPlatform: req.AiPlatform, DateDiscovered: req.DateDiscovered, LastChecked: req.LastChecked,
		Sentiment: sentiment, BrandPositioning: positioning, CitationFrequency: req.CitationFrequency,
		Snippet: req.Snippet, IsCompetitor: req.IsCompetitor, Notes: req.Notes, IsArchived: req.IsArchived,
	}

	if _, err = s.citationRepository.Create(tx, c); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Citation created successfully"}, nil
}

func (s citationService) GetById(id, userId int) (*CitationResponse, error) {
	if err := s.verifyOwnership(id, userId); err != nil {
		return nil, err
	}

	c, err := s.citationRepository.GetById(id)
	if err != nil {
		return nil, errs.NewNotFoundError("citation not found")
	}

	competitors, _ := s.citationRepository.GetCompetitors(id)
	history, _ := s.citationRepository.GetRankingHistory(id)

	data := toCitationData(*c, competitors, history)
	return &CitationResponse{Status: true, Desc: "Get citation successful", Data: data}, nil
}

func (s citationService) Update(id, userId int, req UpdateCitationRequest) (*SimpleResponse, error) {
	if err := s.verifyOwnership(id, userId); err != nil {
		return nil, err
	}

	tx, err := s.citationRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	c := repository.Citation{
		Id: id, BrandId: req.BrandId, Ranking: req.Ranking, Content: req.Content,
		Url: req.Url, AiPlatform: req.AiPlatform, DateDiscovered: req.DateDiscovered, LastChecked: req.LastChecked,
		Sentiment: req.Sentiment, BrandPositioning: req.BrandPositioning, CitationFrequency: req.CitationFrequency,
		Snippet: req.Snippet, IsCompetitor: req.IsCompetitor, Notes: req.Notes, IsArchived: req.IsArchived,
	}

	if err = s.citationRepository.Update(tx, c); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Citation updated successfully"}, nil
}

func (s citationService) Delete(id, userId int) (*SimpleResponse, error) {
	if err := s.verifyOwnership(id, userId); err != nil {
		return nil, err
	}

	tx, err := s.citationRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.citationRepository.Delete(tx, id); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Citation deleted successfully"}, nil
}

func (s citationService) UpdateNotes(id, userId int, notes string) (*SimpleResponse, error) {
	if err := s.verifyOwnership(id, userId); err != nil {
		return nil, err
	}

	tx, err := s.citationRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.citationRepository.UpdateNotes(tx, id, notes); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Notes updated successfully"}, nil
}

func (s citationService) UpdateArchive(id, userId int, isArchived bool) (*SimpleResponse, error) {
	if err := s.verifyOwnership(id, userId); err != nil {
		return nil, err
	}

	tx, err := s.citationRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.citationRepository.UpdateArchive(tx, id, isArchived); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "Archive status updated successfully"}, nil
}
