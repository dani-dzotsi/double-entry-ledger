package middleware

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

//  middleware for application

const APIKeyHeader = "X-API-Key"

type ctxKey string

const (
	loggerKey    ctxKey = "slog_logger"
	requestIDKey ctxKey = "request_id"
)

func ApiKeyMiddleware(expectedKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader(APIKeyHeader)

		if apiKey == "" || subtle.ConstantTimeCompare([]byte(apiKey), []byte(expectedKey)) != 1 {
			logger := GetLogger(c)

			logger.Warn("unauthorized request attempt",
				slog.String("client_ip", c.ClientIP()))

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Next()
	}
}

func RequestIDAndLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		reqID := c.GetHeader("X-REQUEST-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}

		// attach tracking id to outbound response headers
		c.Writer.Header().Set("X-Request-ID", reqID)

		// create a pre-populated logger
		localLogger := slog.Default().With(
			slog.String("request_id", reqID),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
		)

		c.Set(requestIDKey, reqID)
		c.Set(loggerKey, localLogger)

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		logfn := localLogger.Info
		if statusCode >= 500 {
			logfn = localLogger.Error
		} else if statusCode >= 400 {
			logfn = localLogger.Warn
		}

		logfn("http request proceesed", slog.Int("status", statusCode), slog.Duration("latency", latency), slog.String("client_ip", c.ClientIP()))
	}
}

func GetLogger(c *gin.Context) *slog.Logger {
	if val, exists := c.Get(loggerKey); exists {
		if logger, ok := val.(*slog.Logger); ok {
			return logger
		}
	}
	return slog.Default()
}

func GetRequestID(c *gin.Context) string {
	if val, exists := c.Get(requestIDKey); exists {
		if id, ok := val.(string); ok {
			return id
		}
	}
	return ""
}
