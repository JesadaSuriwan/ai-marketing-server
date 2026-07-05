package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type tokenBucket struct {
	tokens   float64
	maxTokens float64
	refillRate float64
	lastRefill time.Time
	mu         sync.Mutex
}

func (b *tokenBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * b.refillRate
	if b.tokens > b.maxTokens {
		b.tokens = b.maxTokens
	}
	b.lastRefill = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

type RateLimitMiddleware interface {
	Limit() gin.HandlerFunc
}

type rateLimitMiddleware struct {
	mu      sync.Mutex
	buckets map[string]*tokenBucket
	// maxTokens is the burst capacity; refillRate is tokens per second
	maxTokens  float64
	refillRate float64
}

func NewRateLimitMiddleware(maxTokens float64, refillRate float64) RateLimitMiddleware {
	return &rateLimitMiddleware{
		buckets:    make(map[string]*tokenBucket),
		maxTokens:  maxTokens,
		refillRate: refillRate,
	}
}

func (m *rateLimitMiddleware) getBucket(ip string) *tokenBucket {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, ok := m.buckets[ip]
	if !ok {
		b = &tokenBucket{
			tokens:     m.maxTokens,
			maxTokens:  m.maxTokens,
			refillRate: m.refillRate,
			lastRefill: time.Now(),
		}
		m.buckets[ip] = b
	}
	return b
}

func (m *rateLimitMiddleware) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		bucket := m.getBucket(ip)
		if !bucket.allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"status": false,
				"desc":   "too many requests, please try again later",
			})
			return
		}
		c.Next()
	}
}
