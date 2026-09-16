package handler

import (
	"net/http"
	"strconv"

	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/pkg/branddomain/service"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type brandDomainHandler struct {
	brandDomainService service.BrandDomainService
}

func NewBrandDomainHandler(brandDomainService service.BrandDomainService) brandDomainHandler {
	return brandDomainHandler{brandDomainService}
}

func (h brandDomainHandler) GetAll(c *gin.Context) {
	brandId, err := strconv.Atoi(c.Query("brand_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid brand_id"))
		return
	}

	result, err := h.brandDomainService.GetAll(brandId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h brandDomainHandler) Create(c *gin.Context) {
	req := service.CreateBrandDomainRequest{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		errs.HandleError(c, errs.NewBadRequestError(err.Error()))
		return
	}

	userId := c.GetInt("userId")
	result, err := h.brandDomainService.Create(userId, req)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h brandDomainHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid id"))
		return
	}

	userId := c.GetInt("userId")
	result, err := h.brandDomainService.Delete(id, userId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
