package handler

import (
	"net/http"
	"strconv"

	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/pkg/recommendation/service"
	"github.com/gin-gonic/gin"
)

type recommendationHandler struct {
	recommendationService service.RecommendationService
}

func NewRecommendationHandler(recommendationService service.RecommendationService) recommendationHandler {
	return recommendationHandler{recommendationService}
}

type generateRequest struct {
	Guidance string `json:"guidance"`
}

func (h recommendationHandler) Generate(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}
	userId, _ := c.Get("userId")

	// Body is optional — guidance is a nice-to-have, not a required field.
	var req generateRequest
	_ = c.ShouldBindJSON(&req)

	result, err := h.recommendationService.Generate(companyId, userId.(int), req.Guidance)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h recommendationHandler) List(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}

	result, err := h.recommendationService.List(companyId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

type updateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h recommendationHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid id"))
		return
	}
	userId, _ := c.Get("userId")

	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid request body"))
		return
	}

	result, err := h.recommendationService.UpdateStatus(id, userId.(int), req.Status)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
