package handler

import (
	"net/http"

	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/pkg/claudeapproval/service"
	"github.com/gin-gonic/gin"
)

type claudeApprovalHandler struct {
	claudeApprovalService service.ClaudeApprovalService
}

func NewClaudeApprovalHandler(claudeApprovalService service.ClaudeApprovalService) claudeApprovalHandler {
	return claudeApprovalHandler{claudeApprovalService}
}

func (h claudeApprovalHandler) Create(c *gin.Context) {
	req := service.CreateClaudeApprovalRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		errs.HandleError(c, errs.NewBadRequestError(err.Error()))
		return
	}
	userId := c.GetInt("userId")

	result, err := h.claudeApprovalService.Create(req.CompanyId, userId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h claudeApprovalHandler) Approve(c *gin.Context) {
	token := c.Param("token")

	result, err := h.claudeApprovalService.Approve(token)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
