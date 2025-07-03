package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"salesforge-assignment/internal/logger"
	"time"
)

const LoggerKey = "logger"

func InjectLogger(log *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(LoggerKey, log)
		c.Next()
	}
}

func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.FromContext(c)

		start := time.Now()
		path := c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			path = path + "?" + c.Request.URL.RawQuery
		}

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		event := log.Info()
		if statusCode >= 500 {
			event = log.Error()
		} else if statusCode >= 400 {
			event = log.Warn()
		}

		event.
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", statusCode).
			Dur("latency", latency).
			Str("ip", c.ClientIP()).
			Msg("Request handled")
	}
}
