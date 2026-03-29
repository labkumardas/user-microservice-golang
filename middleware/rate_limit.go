package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	requests int
	resetAt  time.Time
	mu       sync.Mutex
}

var (
	visitors = make(map[string]*visitor)
	globalMu sync.Mutex
)

// RateLimit limits each IP to `maxReq` requests per `window` duration
func RateLimit(maxReq int, window time.Duration) gin.HandlerFunc {
	// cleanup goroutine
	go func() {
		for {
			time.Sleep(window)
			globalMu.Lock()
			for ip, v := range visitors {
				v.mu.Lock()
				if time.Now().After(v.resetAt) {
					delete(visitors, ip)
				}
				v.mu.Unlock()
			}
			globalMu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		globalMu.Lock()
		v, ok := visitors[ip]
		if !ok {
			v = &visitor{}
			visitors[ip] = v
		}
		globalMu.Unlock()

		v.mu.Lock()
		defer v.mu.Unlock()

		if time.Now().After(v.resetAt) {
			v.requests = 0
			v.resetAt = time.Now().Add(window)
		}

		v.requests++
		if v.requests > maxReq {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "too many requests, please slow down",
			})
			return
		}

		c.Next()
	}
}
