package probe

import (
	"context"
	"sync"
	"time"
)

// Limiter throttles outbound requests to a maximum rate per second.
// A rate of 0 disables throttling.
type Limiter struct {
	rate    int
	tokens  chan struct{}
	stop    chan struct{}
	stopped chan struct{}
	once    sync.Once
}

// NewLimiter creates a rate limiter. rate must be >= 0.
func NewLimiter(rate int) *Limiter {
	l := &Limiter{
		rate:    rate,
		stop:    make(chan struct{}),
		stopped: make(chan struct{}),
	}
	if rate > 0 {
		l.tokens = make(chan struct{}, rate)
		for i := 0; i < rate; i++ {
			l.tokens <- struct{}{}
		}
		go l.run()
	}
	return l
}

func (l *Limiter) run() {
	defer close(l.stopped)
	interval := time.Second / time.Duration(l.rate)
	if interval <= 0 {
		interval = time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			select {
			case l.tokens <- struct{}{}:
			default:
			}
		case <-l.stop:
			return
		}
	}
}

// Wait blocks until a request token is available or ctx is cancelled.
func (l *Limiter) Wait(ctx context.Context) error {
	if l.rate == 0 {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-l.tokens:
		return nil
	}
}

// Stop releases background goroutines.
func (l *Limiter) Stop() {
	l.once.Do(func() {
		if l.rate > 0 {
			close(l.stop)
			<-l.stopped
		}
	})
}
