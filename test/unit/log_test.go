package unit

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"salesforge-assignment/internal/logger"
	"salesforge-assignment/internal/middleware"
	"strings"
	"testing"
	"time"
)

func captureLogsMiddleware(buf *strings.Builder) gin.HandlerFunc {
	writer := zerolog.New(zerolog.ConsoleWriter{Out: buf, NoColor: true})
	l := zerolog.New(writer)
	return middleware.InjectLogger(&l)
}

func TestInjectLogger_SetsLoggerInContext(t *testing.T) {
	// Arrange
	buf := new(strings.Builder)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(captureLogsMiddleware(buf))
	r.GET("/test", func(c *gin.Context) {
		log := logger.FromContext(c)
		log.Info().Msg("hello world")
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	output := buf.String()
	assert.Contains(t, output, "hello world")
}

func TestGinLogger_EmitsStructuredLog(t *testing.T) {
	buf := new(strings.Builder)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(captureLogsMiddleware(buf))
	r.Use(middleware.GinLogger())
	r.GET("/ping", func(c *gin.Context) {
		time.Sleep(10 * time.Millisecond)
		c.String(http.StatusOK, "pong")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping?foo=bar", nil)
	req.RemoteAddr = "192.168.0.1:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.5")
	req.Header.Set("X-Real-IP", "203.0.113.5")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	output := buf.String()

	assert.Contains(t, output, "\"method\":\"GET\"")
	assert.Contains(t, output, "\"path\":\"/ping?foo=bar\"")
	assert.Contains(t, output, "\"status\":200")
	assert.Contains(t, output, "\"latency\":")
	assert.Contains(t, output, "\"ip\":\"203.0.113.5\"")
}
