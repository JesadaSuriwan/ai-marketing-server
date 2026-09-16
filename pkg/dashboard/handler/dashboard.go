package handler

import (
	"net/http"
	"strconv"

	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/pkg/dashboard/service"
	"github.com/gin-gonic/gin"
)

type dashboardHandler struct {
	dashboardService service.DashboardService
}

func NewDashboardHandler(dashboardService service.DashboardService) dashboardHandler {
	return dashboardHandler{dashboardService}
}

func (h dashboardHandler) GetStats(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}

	result, err := h.dashboardService.GetStats(companyId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h dashboardHandler) GetTopPrompts(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}

	result, err := h.dashboardService.GetTopPrompts(companyId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h dashboardHandler) GetTopDomains(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}

	result, err := h.dashboardService.GetTopDomains(companyId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h dashboardHandler) GetRecentCitations(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}

	result, err := h.dashboardService.GetRecentCitations(companyId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h dashboardHandler) GetVisibilityTrend(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}
	interval := c.DefaultQuery("interval", "month")
	result, err := h.dashboardService.GetVisibilityTrend(companyId, interval)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h dashboardHandler) GetPlatformBreakdown(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}
	result, err := h.dashboardService.GetPlatformBreakdown(companyId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h dashboardHandler) GetPromptTrend(c *gin.Context) {
	promptId, err := strconv.Atoi(c.Query("prompt_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid prompt_id"))
		return
	}
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}
	interval := c.DefaultQuery("interval", "month")
	from := c.Query("from")
	to := c.Query("to")

	result, err := h.dashboardService.GetPromptTrend(promptId, companyId, interval, from, to)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h dashboardHandler) GetCompanyMetrics(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}
	result, err := h.dashboardService.GetCompanyMetrics(companyId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h dashboardHandler) GetPromptRankings(c *gin.Context) {
	promptId, err := strconv.Atoi(c.Query("prompt_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid prompt_id"))
		return
	}
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}
	from := c.Query("from")
	to := c.Query("to")
	result, err := h.dashboardService.GetPromptRankings(promptId, companyId, from, to)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h dashboardHandler) GetPromptsOverview(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}
	from := c.Query("from")
	to := c.Query("to")
	result, err := h.dashboardService.GetPromptsOverview(companyId, from, to)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h dashboardHandler) GetBrandRanking(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}
	result, err := h.dashboardService.GetBrandRanking(companyId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h dashboardHandler) GetBrandCitations(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}
	brandId, err := strconv.Atoi(c.Query("brand_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid brand_id"))
		return
	}
	result, err := h.dashboardService.GetBrandCitations(companyId, brandId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h dashboardHandler) GetTopPromptsByBrand(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}
	result, err := h.dashboardService.GetTopPromptsByBrand(companyId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h dashboardHandler) GetTopCitationURLs(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}
	result, err := h.dashboardService.GetTopCitationURLs(companyId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h dashboardHandler) GetCitationURLs(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}
	result, err := h.dashboardService.GetCitationURLs(companyId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h dashboardHandler) GetCitationWinnersLosers(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}
	result, err := h.dashboardService.GetCitationWinnersLosers(companyId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h dashboardHandler) GetCitationURLPrompts(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}
	url := c.Query("url")
	if url == "" {
		errs.HandleError(c, errs.NewBadRequestError("url is required"))
		return
	}
	result, err := h.dashboardService.GetCitationURLPrompts(url, companyId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h dashboardHandler) GetPromptDomains(c *gin.Context) {
	promptId, err := strconv.Atoi(c.Query("prompt_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid prompt_id"))
		return
	}
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}
	from := c.Query("from")
	to := c.Query("to")
	result, err := h.dashboardService.GetPromptDomains(promptId, companyId, from, to)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
