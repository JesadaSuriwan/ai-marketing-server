package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/pkg/visibility/service"
)

type visibilityHandler struct {
	visibilityService service.VisibilityService
}

func NewVisibilityHandler(visibilityService service.VisibilityService) visibilityHandler {
	return visibilityHandler{visibilityService}
}

func (h visibilityHandler) GetAll(c *gin.Context) {
	brandId, err := strconv.Atoi(c.Query("brand_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid brand_id"))
		return
	}

	from := c.Query("from")
	to := c.Query("to")

	result, err := h.visibilityService.GetAll(brandId, from, to)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h visibilityHandler) Create(c *gin.Context) {
	req := service.CreateVisibilityRequest{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		errs.HandleError(c, errs.NewBadRequestError(err.Error()))
		return
	}

	result, err := h.visibilityService.Create(req)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h visibilityHandler) GetBreakdown(c *gin.Context) {
	brandId, err := strconv.Atoi(c.Query("brand_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid brand_id"))
		return
	}

	result, err := h.visibilityService.GetBreakdown(brandId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
