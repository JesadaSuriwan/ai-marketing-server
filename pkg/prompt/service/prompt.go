package service

type PromptData struct {
	Id        int      `json:"id"`
	CompanyId int      `json:"company_id"`
	TagId     *int     `json:"tag_id"`
	TagName   *string  `json:"tag_name"`
	Countries []string `json:"countries"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Active    bool     `json:"active"`
	CreatedAt string   `json:"created_at"`
}

type SetActiveRequest struct {
	Active bool `json:"active"`
}

type CreatePromptRequest struct {
	CompanyId    int      `json:"company_id,string" binding:"required"`
	TagId        *int     `json:"tag_id,string"`
	CountryCodes []string `json:"country_codes"`
	Title        string   `json:"title" binding:"required"`
	Content      string   `json:"content" binding:"required"`
}

// Pointers/nil-slice throughout: a field omitted from the request body is
// left untouched rather than overwritten with an empty value — see
// PromptRepository.Update. CountryCodes specifically: nil (key omitted)
// means "don't touch", a non-nil (possibly empty) slice replaces the full
// set — encoding/json preserves the nil-vs-empty-array distinction, so an
// explicit [] does clear every country.
type UpdatePromptRequest struct {
	TagId        *int     `json:"tag_id,string"`
	CountryCodes []string `json:"country_codes"`
	Title        *string  `json:"title"`
	Content      *string  `json:"content"`
}

type PromptListResponse struct {
	Status bool         `json:"status"`
	Desc   string       `json:"desc"`
	Data   []PromptData `json:"data"`
}

type PromptResponse struct {
	Status bool       `json:"status"`
	Desc   string     `json:"desc"`
	Data   PromptData `json:"data"`
}

type SimpleResponse struct {
	Status bool   `json:"status"`
	Desc   string `json:"desc"`
}

type PromptService interface {
	GetAll(companyId int) (*PromptListResponse, error)
	Create(req CreatePromptRequest) (*PromptResponse, error)
	GetById(id, userId int) (*PromptResponse, error)
	Update(id, userId int, req UpdatePromptRequest) (*SimpleResponse, error)
	SetActive(id, userId int, active bool) (*SimpleResponse, error)
	Delete(id, userId int) (*SimpleResponse, error)
}
