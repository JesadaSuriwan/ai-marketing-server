package handler

import (
	"net/http"
	"strconv"

	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/pkg/promptrun/service"
	"github.com/gin-gonic/gin"
)

type promptRunHandler struct {
	promptRunService service.PromptRunService
}

func NewPromptRunHandler(promptRunService service.PromptRunService) promptRunHandler {
	return promptRunHandler{promptRunService}
}

func (h promptRunHandler) Run(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid id"))
		return
	}
	userId, _ := c.Get("userId")

	result, err := h.promptRunService.Run(id, userId.(int))
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h promptRunHandler) GetHistory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid id"))
		return
	}
	userId, _ := c.Get("userId")

	result, err := h.promptRunService.GetHistory(id, userId.(int))
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h promptRunHandler) GetRunLogs(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}

	filters := service.RunLogFilters{
		From:      c.Query("from"),
		To:        c.Query("to"),
		Engines:   c.Query("engines"),
		TagIds:    c.Query("tag_ids"),
		Countries: c.Query("countries"),
	}
	result, err := h.promptRunService.GetRunLogs(companyId, filters)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
