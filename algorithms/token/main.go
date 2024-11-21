package main

import (
	"fmt"
	"sync"
	"time"
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

func main() {
	tb := NewTokenBucket(2, 5)
	counter := 1

	for i := 0; i < 20; i++ {
		fmt.Printf("[%v] New incoming request %d with algorithm: token bucket\n", time.Now().Format("2006-01-02 15:04:05.999999"), counter)
		if tb.Allow() {
			fmt.Println("Request allowed [v]")
		} else {
			fmt.Println("Request denied [x]")
		}
		time.Sleep(100 * time.Millisecond)
		counter++
		fmt.Println("========================")
	}
}
