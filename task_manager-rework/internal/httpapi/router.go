package httpapi

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/avernus666/task_manager-rework/internal/config"
)

func NewRouter(db *pgxpool.Pool, cfg config.Config, logger *zap.Logger) *gin.Engine {
	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	router.Use(gin.Recovery(), securityHeaders(), requestSizeLimit(1<<20), rateLimit(120, time.Minute), corsMiddleware(), requestLogger(logger))

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Task Manager API", "docs": "API endpoints are listed in README.md"})
	})
	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/health/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := db.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	registerAuthRoutes(router, db, cfg)
	registerProjectRoutes(router, db, cfg)
	registerTaskRoutes(router, db, cfg)
	registerCommentRoutes(router, db, cfg)
	registerNotificationRoutes(router, db, cfg)
	return router
}

func requestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("http request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", time.Since(start)),
		)
	}
}

func abortJSON(c *gin.Context, code int, message string) {
	c.AbortWithStatusJSON(code, gin.H{"error": message})
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowed := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
		if allowed == "" {
			allowed = "http://localhost:5173"
		}
		for _, candidate := range strings.Split(allowed, ",") {
			if origin != "" && origin == strings.TrimSpace(candidate) {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				break
			}
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Header("Cross-Origin-Resource-Policy", "same-site")
		c.Next()
	}
}

func requestSizeLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}

type rateWindow struct {
	start time.Time
	count int
}

func rateLimit(maxRequests int, window time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	clients := make(map[string]rateWindow)
	return func(c *gin.Context) {
		now := time.Now()
		ip := c.ClientIP()
		mu.Lock()
		entry := clients[ip]
		if entry.start.IsZero() || now.Sub(entry.start) >= window {
			entry = rateWindow{start: now}
		}
		entry.count++
		clients[ip] = entry
		if len(clients) > 10000 {
			for key, value := range clients {
				if now.Sub(value.start) >= window {
					delete(clients, key)
				}
			}
		}
		allowed := entry.count <= maxRequests
		mu.Unlock()
		if !allowed {
			c.Header("Retry-After", strconv.Itoa(max(1, int(window.Seconds()))))
			abortJSON(c, http.StatusTooManyRequests, "rate limit exceeded; retry shortly")
			return
		}
		c.Next()
	}
}
