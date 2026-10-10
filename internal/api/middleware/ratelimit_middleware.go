package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type userSession struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	ipLimiter = make(map[string]userSession)
	mu        sync.Mutex
	rateLimit rate.Limit = 10
	burst     int        = 20
)

func getUserLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	var outLimiter *rate.Limiter

	if session, exist := ipLimiter[ip]; !exist {
		newLimiter := userSession{
			limiter:  rate.NewLimiter(rateLimit, burst),
			lastSeen: time.Now(),
		}
		ipLimiter[ip] = newLimiter
		outLimiter = newLimiter.limiter
	} else {
		session.lastSeen = time.Now()
		ipLimiter[ip] = session
		outLimiter = session.limiter
	}
	return outLimiter
}

func RateLimitMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !getUserLimiter(ip).Allow() {
			c.AbortWithStatusJSON(
				http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"},
			)
			return
		}

		c.Next()
	}
}


func CleanupUsers() {
	mu.Lock()
	defer mu.Unlock()
	for ip, sessionLimiter := range ipLimiter {
		if time.Since(sessionLimiter.lastSeen) > 10*time.Minute {
			delete(ipLimiter, ip)
		}
	}
}
