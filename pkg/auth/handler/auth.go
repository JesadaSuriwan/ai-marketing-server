package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/ai-marketing/ai-marketing-server/errs"
	"github.com/ai-marketing/ai-marketing-server/pkg/auth/service"
)

type authHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) authHandler {
	return authHandler{authService}
}

func isSecure() bool {
	return viper.GetString("server.environment") != "dev"
}

func setCookie(c *gin.Context, name, value string, maxAge int) {
	domain := viper.GetString("server.server_domain")
	secure := isSecure()
	c.SetCookie(name, value, maxAge, "/", domain, secure, true)
	if secure {
		c.SetSameSite(http.SameSiteNoneMode)
	} else {
		c.SetSameSite(http.SameSiteLaxMode)
	}
}

func (h authHandler) Register(c *gin.Context) {
	req := service.RegisterRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		errs.HandleError(c, errs.NewBadRequestError(err.Error()))
		return
	}

	result, token, err := h.authService.Register(req)
	if err != nil {
		errs.HandleError(c, err)
		return
	}

	setCookie(c, "Authorization", token, int(24*time.Hour.Seconds()))
	c.JSON(http.StatusCreated, result)
}

func (h authHandler) Login(c *gin.Context) {
	req := service.LoginRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		errs.HandleError(c, errs.NewBadRequestError(err.Error()))
		return
	}

	result, token, err := h.authService.Login(req)
	if err != nil {
		errs.HandleError(c, err)
		return
	}

	setCookie(c, "Authorization", token, int(24*time.Hour.Seconds()))
	c.JSON(http.StatusOK, result)
}

func (h authHandler) Logout(c *gin.Context) {
	setCookie(c, "Authorization", "", -1)
	c.JSON(http.StatusOK, gin.H{"status": true, "desc": "Logged out successfully"})
}

func (h authHandler) Me(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		errs.HandleError(c, errs.NewBadRequestError("userId not found in context"))
		return
	}

	result, err := h.authService.Me(userId.(int))
	if err != nil {
		errs.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
