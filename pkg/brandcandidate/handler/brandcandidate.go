package handler

import (
	"net/http"
	"strconv"

	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/pkg/brandcandidate/service"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type brandCandidateHandler struct {
	brandCandidateService service.BrandCandidateService
}

func NewBrandCandidateHandler(brandCandidateService service.BrandCandidateService) brandCandidateHandler {
	return brandCandidateHandler{brandCandidateService}
}

func (h brandCandidateHandler) List(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}

	result, err := h.brandCandidateService.List(companyId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h brandCandidateHandler) Resolve(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid id"))
		return
	}

	req := service.ResolveRequest{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		errs.HandleError(c, errs.NewBadRequestError(err.Error()))
		return
	}

	result, err := h.brandCandidateService.Resolve(id, req)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
