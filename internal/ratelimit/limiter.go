package ratelimit

import (
	"net"
	"sync"
	"time"
)

type RateLimiter struct {
	mu              sync.Mutex
	limiters        map[string]*tokenBucket
	rate            float64
	burst           int
	cleanupInterval time.Duration
	lastCleanup     time.Time
}

type tokenBucket struct {
	tokens   float64
	lastTime time.Time
}

type Result struct {
	Allowed    bool
	Remaining  float64
	RetryAfter time.Duration
}

func NewRateLimiter(rps float64, burst int, cleanupInterval time.Duration) *RateLimiter {
	return &RateLimiter{
		limiters:        make(map[string]*tokenBucket),
		rate:            rps,
		burst:           burst,
		cleanupInterval: cleanupInterval,
		lastCleanup:     time.Now(),
	}
}

func (rl *RateLimiter) Allow(key string) Result {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	if now.Sub(rl.lastCleanup) > rl.cleanupInterval {
		rl.cleanup(now)
		rl.lastCleanup = now
	}

	bucket, exists := rl.limiters[key]
	if !exists {
		bucket = &tokenBucket{
			tokens:   float64(rl.burst) - 1,
			lastTime: now,
		}
		rl.limiters[key] = bucket
		return Result{
			Allowed:   true,
			Remaining: bucket.tokens,
		}
	}

	elapsed := now.Sub(bucket.lastTime).Seconds()
	bucket.tokens += elapsed * rl.rate
	if bucket.tokens > float64(rl.burst) {
		bucket.tokens = float64(rl.burst)
	}
	bucket.lastTime = now

	if bucket.tokens < 1 {
		secondsNeeded := (1 - bucket.tokens) / rl.rate
		retryAfter := time.Duration(secondsNeeded * float64(time.Second))
		if retryAfter <= 0 {
			retryAfter = time.Millisecond
		}
		return Result{
			Allowed:    false,
			Remaining:  0,
			RetryAfter: retryAfter,
		}
	}

	bucket.tokens -= 1

	return Result{
		Allowed:   true,
		Remaining: bucket.tokens,
	}
}

func (rl *RateLimiter) AllowIP(ip net.IP) Result {
	return rl.Allow(ip.String())
}

func (rl *RateLimiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.limiters, key)
}

func (rl *RateLimiter) cleanup(now time.Time) {
	threshold := now.Add(-rl.cleanupInterval * 2)
	for key, bucket := range rl.limiters {
		if bucket.lastTime.Before(threshold) {
			delete(rl.limiters, key)
		}
	}
}
