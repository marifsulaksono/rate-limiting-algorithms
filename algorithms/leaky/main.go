package main

import (
	"fmt"
	"sync"
	"time"
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
		// jika leak kurang dari 0, set ke 0
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

func main() {
	lb := NewLeakyBucket(5, 2)
	counter := 1

	for i := 0; i < 20; i++ {
		fmt.Printf("[%v] New incoming request %d with algorithm: leaky bucket\n", time.Now().Format("2006-01-02 15:04:05.999999"), counter)
		if lb.Allow() {
			fmt.Println("Request allowed [v]")
		} else {
			fmt.Println("Request denied [x]")
		}
		time.Sleep(100 * time.Millisecond)
		counter++
		fmt.Println("========================")
	}
}
