package service

type ApiKeyData struct {
	Id         int     `json:"id"`
	CompanyId  int     `json:"company_id"`
	Name       string  `json:"name"`
	KeyPrefix  string  `json:"key_prefix"`
	Status     string  `json:"status"`
	LastUsedAt *string `json:"last_used_at"`
	CreatedAt  string  `json:"created_at"`
}

type CreateApiKeyRequest struct {
	CompanyId int    `json:"company_id,string" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type ApiKeyListResponse struct {
	Status bool         `json:"status"`
	Desc   string       `json:"desc"`
	Data   []ApiKeyData `json:"data"`
}

// CreateApiKeyResponse returns the plaintext key once — it cannot be retrieved again.
type CreateApiKeyResponse struct {
	Status   bool   `json:"status"`
	Desc     string `json:"desc"`
	PlainKey string `json:"plain_key"`
}

type SimpleResponse struct {
	Status bool   `json:"status"`
	Desc   string `json:"desc"`
}

type ApiKeyService interface {
	GetAll(companyId int) (*ApiKeyListResponse, error)
	Create(req CreateApiKeyRequest) (*CreateApiKeyResponse, error)
	Revoke(id, userId int) (*SimpleResponse, error)
}
