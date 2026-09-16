package handler

import (
	"net/http"
	"strconv"

	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/pkg/promptsuggestion/service"
	"github.com/gin-gonic/gin"
)

type promptSuggestionHandler struct {
	promptSuggestionService service.PromptSuggestionService
}

func NewPromptSuggestionHandler(promptSuggestionService service.PromptSuggestionService) promptSuggestionHandler {
	return promptSuggestionHandler{promptSuggestionService}
}

func (h promptSuggestionHandler) Generate(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}
	userId, _ := c.Get("userId")

	result, err := h.promptSuggestionService.Generate(companyId, userId.(int))
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h promptSuggestionHandler) List(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}

	result, err := h.promptSuggestionService.List(companyId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

type updateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h promptSuggestionHandler) UpdateStatus(c *gin.Context) {
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

	result, err := h.promptSuggestionService.UpdateStatus(id, userId.(int), req.Status)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
