package unit

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	_ "salesforge-assignment/internal/logger"
	"salesforge-assignment/internal/middleware/auth"
	"testing"
	"time"
)

func dummyLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		l := zerolog.Nop()
		c.Set("logger", &l)
		c.Next()
	}
}

// setupRouter applies dummy logger and auth middleware
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(dummyLoggerMiddleware())
	r.Use(auth.AuthMiddleware())
	return r
}

func TestGenerateToken_Valid(t *testing.T) {
	userID := "user123"
	username := "testuser"
	jwtKey := []byte("mysecretkey")

	tokenString, err := auth.GenerateToken(userID, username, jwtKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	parsed, err := jwt.ParseWithClaims(tokenString, &auth.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	assert.NoError(t, err)
	claims, ok := parsed.Claims.(*auth.Claims)
	assert.True(t, ok)
	assert.Equal(t, userID, claims.UserId)
	assert.Equal(t, username, claims.Username)

	expiresAt := time.Unix(claims.ExpiresAt, 0)
	now := time.Now()
	difference := expiresAt.Sub(now)
	assert.True(t, difference > 59*time.Minute && difference <= time.Hour)
}

func TestAuthMiddleware_NoHeader(t *testing.T) {
	os.Setenv("JWT_SECRET_KEY", "testkey")
	r := setupRouter()
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authorization header is required")
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	os.Setenv("JWT_SECRET_KEY", "anotherkey")
	r := setupRouter()
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.value")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid token")
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	jwtKey := []byte("supersecret")
	os.Setenv("JWT_SECRET_KEY", string(jwtKey))

	tokenStr, err := auth.GenerateToken("u1", "user1", jwtKey)
	assert.NoError(t, err)

	r := setupRouter()
	r.GET("/protected", func(c *gin.Context) {
		claims, exists := c.Get("user_claims")
		assert.True(t, exists)
		c.JSON(http.StatusOK, gin.H{"claims": claims})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "user_id")
	assert.Contains(t, body, "username")
}
