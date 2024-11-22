package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

type LeakyBucket struct {
	capacity int        // kapasitas bucket
	rate     int        // rate leaky (kebocoran) bucket per detik
	queue    int        // jumlah request saat ini
	lastLeak time.Time  // waktu terakhir request leak
	mu       sync.Mutex // mutex untuk menghindari race condition
}

func NewLeakyBucket(capacity, rate int) *LeakyBucket {
	return &LeakyBucket{
		capacity: capacity,
		rate:     rate,
		lastLeak: time.Now(),
	}
}

func (lb *LeakyBucket) Allow() bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(lb.lastLeak).Seconds() // selisih waktu sekarang dengan terakhir diupdate

	// bocorkan request sesuai dengan rate
	leaked := int(elapsed * float64(lb.rate))
	if leaked > 0 {
		fmt.Printf("Leaking %d requests\n", leaked)

		lb.queue -= leaked
		if lb.queue < 0 {
			lb.queue = 0
		}
		lb.lastLeak = now
	}

	fmt.Printf("Bucket state: Queue = %d, Capacity = %d\n", lb.queue, lb.capacity)

	// cek apakah ruang masih tersedia
	if lb.queue < lb.capacity {
		lb.queue++
		return true
	}

	return false
}

// global map. tips: set di dalam middleware jika ingin pemisahan map tiap router
var leakyBuckets = make(map[string]*LeakyBucket)
var leakyMu sync.Mutex

func getLeakyBucket(ip string, capacity, rate int) *LeakyBucket {
	leakyMu.Lock()
	defer leakyMu.Unlock()

	lb, exists := leakyBuckets[ip]
	if !exists {
		lb = NewLeakyBucket(capacity, rate)
		leakyBuckets[ip] = lb
	}
	return lb
}

func LeakyBucketMiddleware(capacity, rate int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			lb := getLeakyBucket(ip, capacity, rate)

			if lb.Allow() {
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
