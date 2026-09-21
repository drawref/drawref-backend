package drawref

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var globalLoginLimiter = rate.NewLimiter(rate.Every(500*time.Millisecond), 10) // ~2 req/sec, burst of 10

func GlobalLoginRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !globalLoginLimiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Login temporarily throttled"})
			return
		}
		c.Next()
	}
}

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	loginLimiters = make(map[string]*clientLimiter)
	loginMu       sync.Mutex
)

func InitRateLimitCleanup() {
	go cleanupStaleLimiters(5*time.Minute, 10*time.Minute)
}

// removes limiters for IPs not seen within maxAge
func cleanupStaleLimiters(interval, maxAge time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		loginMu.Lock()
		now := time.Now()
		for ip, cl := range loginLimiters {
			if now.Sub(cl.lastSeen) > maxAge {
				delete(loginLimiters, ip)
			}
		}
		loginMu.Unlock()
	}
}

func getSubnetKey(rawIP string) string {
	ip := net.ParseIP(rawIP)
	if ip == nil {
		return rawIP
	}
	if v4 := ip.To4(); v4 != nil {
		mask := net.CIDRMask(24, 32)
		return v4.Mask(mask).String()
	}
	mask := net.CIDRMask(64, 128)
	return ip.Mask(mask).String()
}

func PerIPLoginRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		subnet := getSubnetKey(c.ClientIP())

		loginMu.Lock()
		cl, exists := loginLimiters[subnet]
		if !exists {
			cl = &clientLimiter{
				limiter: rate.NewLimiter(rate.Every(3*time.Second), 3),
			}
			loginLimiters[subnet] = cl
		}
		cl.lastSeen = time.Now()
		limiter := cl.limiter
		loginMu.Unlock()

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Wait, you fast bastard"})
			return
		}
		c.Next()
	}
}
