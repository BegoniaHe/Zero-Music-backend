package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Generate New ID", func(t *testing.T) {
		r := gin.New()
		r.Use(RequestID())
		r.GET("/test", func(c *gin.Context) {
			id := GetRequestID(c)
			assert.NotEmpty(t, id)
			assert.Len(t, id, 32)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Header().Get(RequestIDHeader))
	})

	t.Run("Preserve Existing ID", func(t *testing.T) {
		existingID := "existing-id-12345"
		r := gin.New()
		r.Use(RequestID())
		r.GET("/test", func(c *gin.Context) {
			id := GetRequestID(c)
			assert.Equal(t, existingID, id)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set(RequestIDHeader, existingID)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, existingID, w.Header().Get(RequestIDHeader))
	})
}
