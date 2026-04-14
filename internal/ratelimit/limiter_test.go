package ratelimit

import (
	"net"
	"testing"
	"time"
)

func TestRateLimiterAllowsWithinBurst(t *testing.T) {
	rl := NewRateLimiter(10, 5, time.Minute)

	for i := 0; i < 5; i++ {
		result := rl.Allow("test-key")
		if !result.Allowed {
			t.Errorf("request %d should be allowed", i+1)
		}
	}
}

func TestRateLimiterBlocksOverBurst(t *testing.T) {
	rl := NewRateLimiter(1, 2, time.Minute)

	result := rl.Allow("test-key")
	if !result.Allowed {
		t.Error("first request should be allowed")
	}

	result = rl.Allow("test-key")
	if !result.Allowed {
		t.Error("second request should be allowed")
	}

	result = rl.Allow("test-key")
	if result.Allowed {
		t.Error("third request should be denied (burst=2)")
	}

	if result.RetryAfter <= 0 {
		t.Error("RetryAfter should be positive when denied")
	}
}

func TestRateLimiterRefillsTokens(t *testing.T) {
	rl := NewRateLimiter(1000, 1, time.Minute)

	result := rl.Allow("test-key")
	if !result.Allowed {
		t.Error("first request should be allowed")
	}

	result = rl.Allow("test-key")
	if result.Allowed {
		t.Error("second request should be denied (burst=1)")
	}

	time.Sleep(2 * time.Millisecond)

	result = rl.Allow("test-key")
	if !result.Allowed {
		t.Error("request after refill should be allowed")
	}
}

func TestRateLimiterAllowIP(t *testing.T) {
	rl := NewRateLimiter(1, 1, time.Minute)

	ip := net.ParseIP("127.0.0.1")
	result := rl.AllowIP(ip)
	if !result.Allowed {
		t.Error("first IP request should be allowed")
	}

	result = rl.AllowIP(ip)
	if result.Allowed {
		t.Error("second IP request should be denied (burst=1)")
	}

	ip2 := net.ParseIP("192.168.1.1")
	result = rl.AllowIP(ip2)
	if !result.Allowed {
		t.Error("different IP should have its own bucket")
	}
}

func TestRateLimiterReset(t *testing.T) {
	rl := NewRateLimiter(1, 1, time.Minute)

	rl.Allow("test-key")
	rl.Allow("test-key")
	rl.Reset("test-key")

	result := rl.Allow("test-key")
	if !result.Allowed {
		t.Error("request after reset should be allowed")
	}
}

func TestRateLimiterCleanup(t *testing.T) {
	rl := NewRateLimiter(1, 100, 10*time.Millisecond)

	rl.Allow("old-key")
	time.Sleep(30 * time.Millisecond)

	rl.Allow("new-key")
	rl.cleanup(time.Now())

	result := rl.Allow("old-key")
	if !result.Allowed {
		t.Error("old key bucket should have been cleaned up, allowing new requests")
	}
}
