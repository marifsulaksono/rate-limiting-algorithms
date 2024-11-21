package main

import (
	"fmt"
	"sync"
	"time"
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

func main() {
	swl := NewSlidingWindowLogs(3, 5*time.Second)
	counter := 1

	for i := 0; i < 20; i++ {
		fmt.Printf("[%v] New incoming request %d with algorithm: sliding window logs\n", time.Now().Format("2006-01-02 15:04:05.999999"), counter)
		if swl.Allow() {
			fmt.Println("Request allowed")
		} else {
			fmt.Println("Request denied")
		}
		time.Sleep(1 * time.Second)
		counter++
	}
}
