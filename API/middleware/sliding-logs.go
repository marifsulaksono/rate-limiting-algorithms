package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

type SlidingWindowLogs struct {
	limit      int           // jumlah maksimal request dalam jendela waktu
	windowSize time.Duration // ukuran jendela waktu (misalnya 1 menit)
	requests   []time.Time   // kumpulan riwayat waktu request yang telah terjadi
	mu         sync.Mutex    // mutex untuk menghindari race condition
}

func NewSlidingWindowLogs(limit int, windowSize time.Duration) *SlidingWindowLogs {
	return &SlidingWindowLogs{
		limit:      limit,
		windowSize: windowSize,
		requests:   []time.Time{},
	}
}

func (swl *SlidingWindowLogs) Allow() bool {
	swl.mu.Lock()
	defer swl.mu.Unlock()

	now := time.Now()

	// hapus request lama (telah kadaluarsa dari jendela waktu) dan simpan request yang masih belum kadaluwrsa
	var recentRequests []time.Time
	for _, reqTime := range swl.requests {
		// cek apakah waktu request masih dalam jendela waktu, jika iya, tambahkan ke recentRequests
		if now.Sub(reqTime) < swl.windowSize {
			recentRequests = append(recentRequests, reqTime)
		} else {
			fmt.Printf("Request at [%v] has expired and was discarded\n", reqTime.Format("2006-01-02 15:04:05.999999"))
		}
	}
	swl.requests = recentRequests

	fmt.Printf("Current window requests: %d, Remaining capacity: %d\n", len(swl.requests), swl.limit-len(swl.requests))

	// cek apakah jumlah request dalam jendela waktu kurang dari limit
	if len(swl.requests) < swl.limit {
		swl.requests = append(swl.requests, now) // simpan request baru
		return true
	}

	return false
}

// global map. tips: set di dalam middleware jika ingin pemisahan map tiap router
var slidingWindowLogs = make(map[string]*SlidingWindowLogs)
var slidingWindowMu sync.Mutex

func getSlidingWindowLog(ip string, limit int, windowSize time.Duration) *SlidingWindowLogs {
	slidingWindowMu.Lock()
	defer slidingWindowMu.Unlock()

	swl, exists := slidingWindowLogs[ip]
	if !exists {
		swl = NewSlidingWindowLogs(limit, windowSize)
		slidingWindowLogs[ip] = swl
	}
	return swl
}

func SlidingWindowMiddleware(limit int, windowSize time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			swl := getSlidingWindowLog(ip, limit, windowSize)

			if swl.Allow() {
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
