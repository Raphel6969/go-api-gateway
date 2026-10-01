package ratelimit

import (
	"context"
	"sync"
	"time"
)

type Bucket struct {
	capacity   float64
	tokens     float64
	refillRate float64
	lastRefill time.Time
	mu         sync.Mutex
}

type RateLimiter struct {
	mu         sync.Mutex
	buckets    map[string]*Bucket
	capacity   float64
	refillRate float64
}

func NewRateLimiter(capacity float64, refillRate float64) *RateLimiter {
	rl := &RateLimiter{
		buckets:    make(map[string]*Bucket),
		capacity:   capacity,
		refillRate: refillRate,
	}

	go rl.cleanupLoop()

	return rl
}

func (rl *RateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	rl.mu.Lock()
	b, exists := rl.buckets[key]
	if !exists {
		b = &Bucket{
			capacity:   rl.capacity,
			tokens:     rl.capacity,
			refillRate: rl.refillRate,
			lastRefill: time.Now(),
		}
		rl.buckets[key] = b
	}
	rl.mu.Unlock()

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.lastRefill = now

	b.tokens += elapsed * b.refillRate
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}

	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return true, nil
	}

	return false, nil
}

func (rl *RateLimiter) AllowSimple(key string) bool {
	allowed, _ := rl.Allow(context.Background(), key)
	return allowed
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		for key, b := range rl.buckets {
			b.mu.Lock()
			if time.Since(b.lastRefill) > 10*time.Minute {
				delete(rl.buckets, key)
			}
			b.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}
