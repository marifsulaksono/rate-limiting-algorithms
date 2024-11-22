package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

type TokenBucket struct {
	rate       float64    // rate refill token ke dalam bucket
	burst      int        // jumlah maksimum token didalam bucket
	tokens     float64    // jumlah token saat ini
	lastUpdate time.Time  // waktu terakhir token diupdate
	mu         sync.Mutex // mutex untuk menghindari race condition
}

func NewTokenBucket(rate float64, burst int) *TokenBucket {
	return &TokenBucket{
		rate:       rate,
		burst:      burst,
		tokens:     float64(burst),
		lastUpdate: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastUpdate).Seconds() // selisih waktu sekarang dengan terakhir diupdate

	// refill token sesuai dengan rate
	tb.tokens += elapsed * tb.rate

	// jika token melebihi batas, set token ke batas (buang token yang melebihi batas)
	if tb.tokens > float64(tb.burst) {
		tb.tokens = float64(tb.burst)
	}
	tb.lastUpdate = now
	fmt.Printf("Tokens available: %.2f token\n", tb.tokens)

	// cek apakah token masih ada
	if tb.tokens >= 1 {
		tb.tokens-- // kurangi token untuk sebuah request
		return true
	}
	return false
}

// global map. tips: set di dalam middleware jika ingin pemisahan map tiap router
var tokenBuckets = make(map[string]*TokenBucket)
var tokenMu sync.Mutex

func getTokenBucket(ip string, rate float64, burst int) *TokenBucket {
	tokenMu.Lock()
	defer tokenMu.Unlock()

	tb, exists := tokenBuckets[ip]
	if !exists {
		tb = NewTokenBucket(rate, burst)
		tokenBuckets[ip] = tb
	}
	return tb
}

func TokenBucketMiddleware(rate float64, burst int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			tb := getTokenBucket(ip, rate, burst)

			if tb.Allow() {
				fmt.Println("Request allowed")
				return next(c)
			}

			fmt.Println("Request denied")
			return c.JSON(http.StatusTooManyRequests, map[string]string{"message": "Terlalu banyak permintaan, coba lagi nanti"})
		}
	}
}
