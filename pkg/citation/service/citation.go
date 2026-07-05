package service

type CompetitorDTO struct {
	Id             int    `json:"id"`
	CompetitorName string `json:"competitor_name"`
}

type RankingHistoryDTO struct {
	Id        int    `json:"id"`
	DateLabel string `json:"date_label"`
	Rank      int    `json:"rank"`
}

type CitationData struct {
	Id                int                 `json:"id"`
	PromptId          int                 `json:"prompt_id"`
	BrandId           *int                `json:"brand_id"`
	Ranking           int                 `json:"ranking"`
	Content           string              `json:"content"`
	Url               *string             `json:"url"`
	AiPlatform        *string             `json:"ai_platform"`
	DateDiscovered    *string             `json:"date_discovered"`
	LastChecked       *string             `json:"last_checked"`
	Sentiment         string              `json:"sentiment"`
	BrandPositioning  string              `json:"brand_positioning"`
	CitationFrequency int                 `json:"citation_frequency"`
	Snippet           *string             `json:"snippet"`
	IsCompetitor      bool                `json:"is_competitor"`
	Notes             *string             `json:"notes"`
	IsArchived        bool                `json:"is_archived"`
	Competitors       []CompetitorDTO     `json:"competitors"`
	RankingHistory    []RankingHistoryDTO `json:"ranking_history"`
	CreatedAt         string              `json:"created_at"`
}

type CreateCitationRequest struct {
	PromptId          int     `json:"prompt_id" binding:"required"`
	BrandId           *int    `json:"brand_id"`
	Ranking           int     `json:"ranking"`
	Content           string  `json:"content" binding:"required"`
	Url               *string `json:"url"`
	AiPlatform        *string `json:"ai_platform"`
	DateDiscovered    *string `json:"date_discovered"`
	LastChecked       *string `json:"last_checked"`
	Sentiment         string  `json:"sentiment"`
	BrandPositioning  string  `json:"brand_positioning"`
	CitationFrequency int     `json:"citation_frequency"`
	Snippet           *string `json:"snippet"`
	IsCompetitor      bool    `json:"is_competitor"`
	Notes             *string `json:"notes"`
	IsArchived        bool    `json:"is_archived"`
}

type UpdateCitationRequest struct {
	BrandId           *int    `json:"brand_id"`
	Ranking           int     `json:"ranking"`
	Content           string  `json:"content"`
	Url               *string `json:"url"`
	AiPlatform        *string `json:"ai_platform"`
	DateDiscovered    *string `json:"date_discovered"`
	LastChecked       *string `json:"last_checked"`
	Sentiment         string  `json:"sentiment"`
	BrandPositioning  string  `json:"brand_positioning"`
	CitationFrequency int     `json:"citation_frequency"`
	Snippet           *string `json:"snippet"`
	IsCompetitor      bool    `json:"is_competitor"`
	Notes             *string `json:"notes"`
	IsArchived        bool    `json:"is_archived"`
}

type UpdateNotesRequest struct {
	Notes string `json:"notes" binding:"required"`
}

type UpdateArchiveRequest struct {
	IsArchived bool `json:"is_archived"`
}

type CitationListResponse struct {
	Status bool           `json:"status"`
	Desc   string         `json:"desc"`
	Data   []CitationData `json:"data"`
}

type CitationResponse struct {
	Status bool         `json:"status"`
	Desc   string       `json:"desc"`
	Data   CitationData `json:"data"`
}

type SimpleResponse struct {
	Status bool   `json:"status"`
	Desc   string `json:"desc"`
}

type CitationService interface {
	GetAll(promptId int) (*CitationListResponse, error)
	Create(req CreateCitationRequest) (*SimpleResponse, error)
	GetById(id, userId int) (*CitationResponse, error)
	Update(id, userId int, req UpdateCitationRequest) (*SimpleResponse, error)
	Delete(id, userId int) (*SimpleResponse, error)
	UpdateNotes(id, userId int, notes string) (*SimpleResponse, error)
	UpdateArchive(id, userId int, isArchived bool) (*SimpleResponse, error)
}
