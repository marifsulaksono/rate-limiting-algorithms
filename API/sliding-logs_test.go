package main

import (
	"net/http"
	"net/http/httptest"
	"rate-limiting/middleware"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestSlidingWindowMiddleware(t *testing.T) {
	e := echo.New()

	e.Use(middleware.SlidingWindowMiddleware(3, 5*time.Second))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Request allowed!")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code) // denied karena terlalu banyak permintaan

	time.Sleep(6 * time.Second)

	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code) // allowed karena beberapa window telah kadaluarsa
}
