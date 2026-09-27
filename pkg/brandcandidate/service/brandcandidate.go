package service

type BrandCandidateData struct {
	Id           int    `json:"id"`
	CompanyId    int    `json:"company_id"`
	Name         string `json:"name"`
	MentionCount int    `json:"mention_count"`
	FirstSeen    string `json:"first_seen"`
	LastSeen     string `json:"last_seen"`
}

type BrandCandidateListResponse struct {
	Status bool                 `json:"status"`
	Desc   string               `json:"desc"`
	Data   []BrandCandidateData `json:"data"`
}

// ResolveRequest — Action is one of:
//   - "own_alias": records Name as another name for the company's own brand
//   - "merge_competitor": records Name as another name for TargetBrandId
//     (must be an existing competitor of this company)
//   - "new_competitor": creates a new competitor brand called Name
//   - "dismiss": not a real brand, just stop suggesting it
type ResolveRequest struct {
	CompanyId     int    `json:"company_id,string" binding:"required"`
	Action        string `json:"action" binding:"required"`
	TargetBrandId *int   `json:"target_brand_id"`
}

type SimpleResponse struct {
	Status bool   `json:"status"`
	Desc   string `json:"desc"`
}

type BrandCandidateService interface {
	List(companyId int) (*BrandCandidateListResponse, error)
	Resolve(candidateId int, req ResolveRequest) (*SimpleResponse, error)
}
