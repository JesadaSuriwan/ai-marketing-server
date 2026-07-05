package middlewares

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	limits "github.com/gin-contrib/size"
	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type ServerMiddleware interface {
	GetCorsMiddleware() gin.HandlerFunc
	GetBodyLimitMiddleware(int) gin.HandlerFunc
	GetTimeoutMiddleware(time.Duration) gin.HandlerFunc
}

type serverMiddleware struct{}

func NewServerMiddleware() ServerMiddleware {
	return serverMiddleware{}
}

func (m serverMiddleware) GetCorsMiddleware() gin.HandlerFunc {
	rawOrigins := viper.GetString("server.client_url")
	allowedOrigins := strings.Split(rawOrigins, ",")
	for i := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
	}
	cfg := cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"X-Requested-With", "Authorization", "Origin", "Content-Length", "Content-Type", "TransactionID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	return cors.New(cfg)
}

func (m serverMiddleware) GetBodyLimitMiddleware(bodyLimit int) gin.HandlerFunc {
	bodyLimitByte := int64(bodyLimit * 1024 * 1024)
	return limits.RequestSizeLimiter(bodyLimitByte)
}

func (m serverMiddleware) GetTimeoutMiddleware(timeoutValue time.Duration) gin.HandlerFunc {
	return timeout.New(
		timeout.WithTimeout(timeoutValue*time.Second),
		timeout.WithResponse(func(c *gin.Context) {
			log.Println("Timeout middleware: Request timed out")
			c.JSON(http.StatusRequestTimeout, gin.H{"status": false, "desc": "timeout"})
		}),
	)
}
