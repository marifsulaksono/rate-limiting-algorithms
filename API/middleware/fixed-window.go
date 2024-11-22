package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

type FixedWindowCounter struct {
	limit      int           // jumlah maksimal request dalam jendela waktu
	windowSize time.Duration // ukuran jendela waktu (contoh 1 menit)
	count      int           // jumlah request yang sudah dilakukan dalam jendela waktu
	resetTime  time.Time     // waktu reset jendela waktu
	mu         sync.Mutex    // mutex untuk menghindari race condition
}

func NewFixedWindowCounter(limit int, windowSize time.Duration) *FixedWindowCounter {
	return &FixedWindowCounter{
		limit:      limit,
		windowSize: windowSize,
		count:      0,
		resetTime:  time.Now().Add(windowSize),
	}
}

func (fwc *FixedWindowCounter) Allow() bool {
	fwc.mu.Lock()
	defer fwc.mu.Unlock()

	now := time.Now()

	// cek apabila waktu sekarang melebihi waktu reset, maka reset jendela waktu
	if now.After(fwc.resetTime) {
		fwc.count = 0
		fwc.resetTime = now.Add(fwc.windowSize) // set waktu reset jendela waktu sesuai dengan waktu sekarang + windowSize
	}

	fmt.Printf("Current requests: %d, Remaining capacity: %d\n", fwc.count, fwc.limit-fwc.count)

	// cek apakah jumlah request dalam jendela waktu kurang dari limit
	if fwc.count < fwc.limit {
		fwc.count++
		return true
	}

	return false
}

// global map. tips: set di dalam middleware jika ingin pemisahan map tiap router
var fixedWindowCounters = make(map[string]*FixedWindowCounter)
var fixedWindowMu sync.Mutex

func getFixedWindowCounter(ip string, limit int, windowSize time.Duration) *FixedWindowCounter {
	fixedWindowMu.Lock()
	defer fixedWindowMu.Unlock()

	fwc, exists := fixedWindowCounters[ip]
	if !exists {
		fwc = NewFixedWindowCounter(limit, windowSize)
		fixedWindowCounters[ip] = fwc
	}
	return fwc
}

func FixedWindowMiddleware(limit int, windowSize time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			fwc := getFixedWindowCounter(ip, limit, windowSize)

			if fwc.Allow() {
				fmt.Println("Request allowed")
				return next(c)
			}

			fmt.Println("Request denied")
			return c.JSON(http.StatusTooManyRequests, map[string]string{
				"message": "Terlalu banyak permintaan, coba lagi nanti",
			})
		}
	}
}
