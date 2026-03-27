package ratelimit

import (
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

type userLimiter struct {
	minute *rate.Limiter
	daily  atomic.Int64
}

type RateLimiter struct {
	mu       sync.Mutex
	limiters map[int64]*userLimiter
	perMin   rate.Limit
	burst    int
	perDay   int64
}

func NewRateLimiter(perMinute, burst, perDay int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[int64]*userLimiter),
		perMin:   rate.Limit(float64(perMinute) / 60.0),
		burst:    burst,
		perDay:   int64(perDay),
	}

	go rl.resetDailyLoop()
	return rl
}

func (rl *RateLimiter) Allow(userID int64) bool {
	rl.mu.Lock()
	ul, ok := rl.limiters[userID]
	if !ok {
		ul = &userLimiter{minute: rate.NewLimiter(rl.perMin, rl.burst)}
		rl.limiters[userID] = ul
	}
	rl.mu.Unlock()

	if ul.daily.Load() >= rl.perDay {
		return false
	}

	if !ul.minute.Allow() {
		return false
	}

	ul.daily.Add(1)
	return true
}

func (rl *RateLimiter) resetDailyLoop() {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		time.Sleep(time.Until(next))

		rl.mu.Lock()
		for _, ul := range rl.limiters {
			ul.daily.Store(0)
		}
		rl.mu.Unlock()
	}
}
