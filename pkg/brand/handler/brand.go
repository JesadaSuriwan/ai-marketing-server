package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/pkg/brand/service"
)

type brandHandler struct {
	brandService service.BrandService
}

func NewBrandHandler(brandService service.BrandService) brandHandler {
	return brandHandler{brandService}
}

func (h brandHandler) GetAll(c *gin.Context) {
	companyId, err := strconv.Atoi(c.Query("company_id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid company_id"))
		return
	}

	result, err := h.brandService.GetAll(companyId)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h brandHandler) Create(c *gin.Context) {
	req := service.CreateBrandRequest{}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		errs.HandleError(c, errs.NewBadRequestError(err.Error()))
		return
	}

	result, err := h.brandService.Create(req)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h brandHandler) GetById(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid id"))
		return
	}
	userId, _ := c.Get("userId")

	result, err := h.brandService.GetById(id, userId.(int))
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h brandHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid id"))
		return
	}
	userId, _ := c.Get("userId")

	req := service.UpdateBrandRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		errs.HandleError(c, errs.NewBadRequestError(err.Error()))
		return
	}

	result, err := h.brandService.Update(id, userId.(int), req)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h brandHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid id"))
		return
	}
	userId, _ := c.Get("userId")

	result, err := h.brandService.Delete(id, userId.(int))
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h brandHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		errs.HandleError(c, errs.NewBadRequestError("invalid id"))
		return
	}
	userId, _ := c.Get("userId")

	req := service.UpdateStatusRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		errs.HandleError(c, errs.NewBadRequestError(err.Error()))
		return
	}

	result, err := h.brandService.UpdateStatus(id, userId.(int), req.Status)
	if err != nil {
		errs.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
