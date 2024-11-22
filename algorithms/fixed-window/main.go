package main

import (
	"fmt"
	"sync"
	"time"
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

func main() {
	fwc := NewFixedWindowCounter(7, 10*time.Second)
	counter := 1

	for i := 0; i < 12; i++ {
		fmt.Printf("[%v] New incoming request %d with algorithm: fixed window counter\n", time.Now().Format("2006-01-02 15:04:05.999999"), counter)
		if fwc.Allow() {
			fmt.Println("Request allowed [v]")
		} else {
			fmt.Println("Request denied [x]")
		}
		time.Sleep(1 * time.Second)
		counter++
	}
}
