package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ai-marketing/ai-marketing-server/logs"
	"github.com/ai-marketing/ai-marketing-server/utils"
)

type AuthMiddleware interface {
	AuthRequired(c *gin.Context)
}

type authMiddleware struct{}

type unauthorizedResponse struct {
	Status bool   `json:"status"`
	Desc   string `json:"desc"`
}

func NewAuthMiddleware() AuthMiddleware {
	return authMiddleware{}
}

func (m authMiddleware) AuthRequired(c *gin.Context) {
	token, err := c.Cookie("Authorization")
	if err != nil {
		logs.Error(err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, unauthorizedResponse{Status: false, Desc: "Unauthorized"})
		return
	}

	userId, role, err := utils.VerifyToken(token)
	if err != nil {
		logs.Error(err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, unauthorizedResponse{Status: false, Desc: "Unauthorized"})
		return
	}

	c.Set("userId", userId)
	c.Set("role", role)
	c.Next()
}
