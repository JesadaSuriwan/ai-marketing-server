package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/pkg/api_key/service"
)

type apiKeyHandler struct {
	apiKeyService service.ApiKeyService
}

func NewApiKeyHandler(apiKeyService service.ApiKeyService) apiKeyHandler {
	return apiKeyHandler{apiKeyService}
}

func (h apiKeyHandler) GetAll(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}

	result, err := h.apiKeyService.GetAll(companyId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h apiKeyHandler) Create(c *gin.Context) {
	req := service.CreateApiKeyRequest{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		errs.HandleError(c, errs.NewBadRequestError(err.Error()))
		return
	}

	result, err := h.apiKeyService.Create(req)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h apiKeyHandler) Revoke(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid id"))
		return
	}

	userId := c.GetInt("userId")
	result, err := h.apiKeyService.Revoke(id, userId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
