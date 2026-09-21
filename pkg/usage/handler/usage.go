package handler

import (
	"net/http"
	"strconv"

	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/pkg/usage/service"
	"github.com/gin-gonic/gin"
)

type usageHandler struct {
	usageService service.UsageService
}

func NewUsageHandler(usageService service.UsageService) usageHandler {
	return usageHandler{usageService}
}

func (h usageHandler) GetSummary(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}
	from := c.Query("from")
	to := c.Query("to")
	result, err := h.usageService.GetSummaryForCompany(companyId, from, to)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
