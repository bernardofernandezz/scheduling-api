package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// SecurityHeaders adds security headers to all responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add security headers to prevent common web vulnerabilities
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Writer.Header().Set("Content-Security-Policy", "default-src 'self'")
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Writer.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		
		// Only add HSTS header in production
		if gin.Mode() == gin.ReleaseMode {
			c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		c.Next()
	}
}

// Client represents a client for rate limiting
type Client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// ClientMap is a thread-safe map of clients for rate limiting
type ClientMap struct {
	clients map[string]*Client
	mu      sync.Mutex
}

// Global client map for rate limiting
var (
	clientMap = &ClientMap{
		clients: make(map[string]*Client),
	}
)

// init starts the cleanup goroutine for the client map
func init() {
	go clientMap.cleanup()
}

// cleanup periodically removes old clients from the map
func (cm *ClientMap) cleanup() {
	for {
		time.Sleep(time.Minute)
		
		cm.mu.Lock()
		for ip, client := range cm.clients {
			// Remove clients that haven't been seen in 3 minutes
			if time.Since(client.lastSeen) > 3*time.Minute {
				delete(cm.clients, ip)
			}
		}
		cm.mu.Unlock()
	}
}

// getClient gets or creates a client for rate limiting
func (cm *ClientMap) getClient(ip string, rps rate.Limit, burst int) *rate.Limiter {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	client, exists := cm.clients[ip]
	if !exists {
		// Create a new client if it doesn't exist
		client = &Client{
			limiter:  rate.NewLimiter(rps, burst),
			lastSeen: time.Now(),
		}
		cm.clients[ip] = client
	} else {
		// Update the last seen time
		client.lastSeen = time.Now()
	}
	
	return client.limiter
}

// RateLimiter creates a middleware that limits request rate per client IP
func RateLimiter(requestsPerMinute int, per time.Duration) gin.HandlerFunc {
	// Calculate requests per second
	rps := rate.Limit(float64(requestsPerMinute) / per.Seconds())
	
	return func(c *gin.Context) {
		// Get client IP
		ip := c.ClientIP()
		
		// Get client limiter
		limiter := clientMap.getClient(ip, rps, requestsPerMinute)
		
		// Check if the request can be processed
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// RequestLogger logs request details
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		
		// Process request
		c.Next()
		
		// Calculate request duration
		duration := time.Since(start)
		
		// Log request
		if gin.Mode() != gin.ReleaseMode {
			// Only log detailed info in non-release mode
			c.JSON(http.StatusOK, gin.H{
				"method":   c.Request.Method,
				"path":     c.Request.URL.Path,
				"status":   c.Writer.Status(),
				"duration": duration.String(),
				"client":   c.ClientIP(),
			})
		}
	}
}

