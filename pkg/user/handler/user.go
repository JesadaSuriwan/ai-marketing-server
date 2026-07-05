package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/pkg/user/service"
)

type userHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) userHandler {
	return userHandler{userService}
}

func (h userHandler) GetById(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid id"))
		return
	}

	result, err := h.userService.GetById(id)
	if err != nil {
		errs.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h userHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid id"))
		return
	}

	req := service.UpdateUserRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		errs.HandleError(c, errs.NewBadRequestError(err.Error()))
		return
	}

	result, err := h.userService.Update(id, req)
	if err != nil {
		errs.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
