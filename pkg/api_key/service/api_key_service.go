package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/pkg/api_key/repository"
)

type apiKeyService struct {
	apiKeyRepository repository.ApiKeyRepository
}

func NewApiKeyService(apiKeyRepository repository.ApiKeyRepository) ApiKeyService {
	return apiKeyService{apiKeyRepository}
}

func generateApiKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return "ak_" + hex.EncodeToString(b)
}

func hashKey(plainKey string) string {
	h := sha256.Sum256([]byte(plainKey))
	return hex.EncodeToString(h[:])
}

func (s apiKeyService) verifyOwnership(id, userId int) error {
	ok, err := s.apiKeyRepository.BelongsToUser(id, userId)
	if err != nil {
		logs.Error(err)
		return errs.NewUnexpectedError()
	}
	if !ok {
		return errs.NewForbiddenError("access denied")
	}
	return nil
}

func (s apiKeyService) GetAll(companyId int) (*ApiKeyListResponse, error) {
	keys, err := s.apiKeyRepository.GetAll(companyId)
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	data := []ApiKeyData{}
	for _, k := range keys {
		data = append(data, ApiKeyData{
			Id: k.Id, CompanyId: k.CompanyId, Name: k.Name,
			KeyPrefix:  k.KeyPrefix,
			Status:     k.Status,
			LastUsedAt: k.LastUsedAt, CreatedAt: k.CreatedAt,
		})
	}

	return &ApiKeyListResponse{Status: true, Desc: "Get API keys successful", Data: data}, nil
}

func (s apiKeyService) Create(req CreateApiKeyRequest) (*CreateApiKeyResponse, error) {
	tx, err := s.apiKeyRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	plainKey := generateApiKey()
	k := repository.ApiKey{
		CompanyId: req.CompanyId,
		Name:      req.Name,
		KeyPrefix: plainKey[:8],
		KeyValue:  hashKey(plainKey),
		Status:    "Active",
	}

	if _, err = s.apiKeyRepository.Create(tx, k); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &CreateApiKeyResponse{
		Status:   true,
		Desc:     "API key created successfully — save this key now, it will not be shown again",
		PlainKey: plainKey,
	}, nil
}

func (s apiKeyService) Revoke(id, userId int) (*SimpleResponse, error) {
	if err := s.verifyOwnership(id, userId); err != nil {
		return nil, err
	}

	tx, err := s.apiKeyRepository.NewTransaction()
	if err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}
	defer tx.Rollback()

	if err = s.apiKeyRepository.Delete(tx, id); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	if err = tx.Commit(); err != nil {
		logs.Error(err)
		return nil, errs.NewUnexpectedError()
	}

	return &SimpleResponse{Status: true, Desc: "API key revoked successfully"}, nil
}
