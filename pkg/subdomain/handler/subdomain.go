package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/pkg/subdomain/service"
)

type subdomainHandler struct {
	subdomainService service.SubdomainService
}

func NewSubdomainHandler(subdomainService service.SubdomainService) subdomainHandler {
	return subdomainHandler{subdomainService}
}

func (h subdomainHandler) GetAll(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}

	result, err := h.subdomainService.GetAll(companyId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h subdomainHandler) Create(c *gin.Context) {
	req := service.CreateSubdomainRequest{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		errs.HandleError(c, errs.NewBadRequestError(err.Error()))
		return
	}

	result, err := h.subdomainService.Create(req)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h subdomainHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid id"))
		return
	}

	userId := c.GetInt("userId")
	result, err := h.subdomainService.Delete(id, userId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
