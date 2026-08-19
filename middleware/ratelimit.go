package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"go-ecommerce/models"

	"github.com/gin-gonic/gin"
)

// RateLimit throttles requests per client IP using a token bucket.
//
// Applied to the auth endpoints, this is what makes an online password-guessing
// attack impractical. Per-account lockout alone is not enough: it does nothing
// against credential stuffing, where an attacker tries one password against
// thousands of different accounts and never trips a single account's counter.
//
// The state is in-process, so each replica enforces its own quota. Behind a load
// balancer or across multiple instances, move this to Redis.
func RateLimit(requests int, window time.Duration) gin.HandlerFunc {
	limiter := &ipRateLimiter{
		buckets:  make(map[string]*bucket),
		capacity: float64(requests),
		refill:   float64(requests) / window.Seconds(),
		idleTTL:  window * 10,
	}

	return func(c *gin.Context) {
		// c.ClientIP() honours X-Forwarded-For, which a client can forge unless
		// Gin is told which proxies to trust. Call router.SetTrustedProxies()
		// before relying on this in production.
		if !limiter.allow(c.ClientIP()) {
			retryAfter := int(window.Seconds())
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, models.ErrorResponse{
				Error: "too many requests, please try again later",
			})
			return
		}
		c.Next()
	}
}

// bucket is one client's token allowance.
type bucket struct {
	tokens   float64
	lastSeen time.Time
}

type ipRateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	capacity float64 // maximum burst
	refill   float64 // tokens restored per second
	idleTTL  time.Duration
	lastGC   time.Time
}

// allow consumes a token for ip, reporting whether the request may proceed.
func (l *ipRateLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	l.collectIdle(now)

	b, exists := l.buckets[ip]
	if !exists {
		l.buckets[ip] = &bucket{tokens: l.capacity - 1, lastSeen: now}
		return true
	}

	// Refill proportionally to the time elapsed, capped at the burst size.
	b.tokens += now.Sub(b.lastSeen).Seconds() * l.refill
	if b.tokens > l.capacity {
		b.tokens = l.capacity
	}
	b.lastSeen = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// collectIdle drops buckets nobody has touched recently, so a long-running
// server does not accumulate a map entry for every IP it has ever seen.
// Callers must hold the mutex.
func (l *ipRateLimiter) collectIdle(now time.Time) {
	if now.Sub(l.lastGC) < l.idleTTL {
		return
	}
	l.lastGC = now
	for ip, b := range l.buckets {
		if now.Sub(b.lastSeen) > l.idleTTL {
			delete(l.buckets, ip)
		}
	}
}
