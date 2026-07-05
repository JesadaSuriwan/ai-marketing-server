package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/pkg/member/service"
)

type memberHandler struct {
	memberService service.MemberService
}

func NewMemberHandler(memberService service.MemberService) memberHandler {
	return memberHandler{memberService}
}

func (h memberHandler) GetAll(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}

	result, err := h.memberService.GetAll(companyId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h memberHandler) Create(c *gin.Context) {
	req := service.CreateMemberRequest{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		errs.HandleError(c, errs.NewBadRequestError(err.Error()))
		return
	}

	result, err := h.memberService.Create(req)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h memberHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid id"))
		return
	}

	userId := c.GetInt("userId")
	result, err := h.memberService.Delete(id, userId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
