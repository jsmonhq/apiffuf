package probe

import (
	"context"
	"testing"
	"time"
)

func TestLimiterRespectsRate(t *testing.T) {
	const rate = 5
	limiter := NewLimiter(rate)
	defer limiter.Stop()

	ctx := context.Background()
	start := time.Now()
	for i := 0; i < rate; i++ {
		if err := limiter.Wait(ctx); err != nil {
			t.Fatalf("Wait() error = %v", err)
		}
	}
	elapsed := time.Since(start)
	if elapsed > 200*time.Millisecond {
		t.Fatalf("first %d tokens took too long: %v", rate, elapsed)
	}

	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("Wait() error = %v", err)
	}
	if elapsed := time.Since(start); elapsed < 150*time.Millisecond {
		t.Fatalf("expected throttle after bucket drained, elapsed=%v", elapsed)
	}
}

func TestLimiterZeroRateIsNoop(t *testing.T) {
	limiter := NewLimiter(0)
	defer limiter.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	start := time.Now()
	for i := 0; i < 20; i++ {
		if err := limiter.Wait(ctx); err != nil {
			t.Fatalf("Wait() error = %v", err)
		}
	}
	if time.Since(start) > 50*time.Millisecond {
		t.Fatalf("zero rate limiter should not throttle")
	}
}
