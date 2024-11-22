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

func TestLeakyBucketMiddleware(t *testing.T) {
	e := echo.New()

	e.Use(middleware.LeakyBucketMiddleware(5, 2))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Request allowed!")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code) // denied karena terlalu banyak permintaan

	time.Sleep(3 * time.Second)

	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code) // allowed karena token telah kosong
}
