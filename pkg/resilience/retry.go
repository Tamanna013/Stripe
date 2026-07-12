// Package resilience implements FlowGuard's standard retry and circuit
// breaker policies, per the architecture spec's resilience patterns.
package resilience

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RetryConfig configures the standard FlowGuard retry envelope.
type RetryConfig struct {
	MaxAttempts       int
	BaseDelay         time.Duration
	MaxDelay          time.Duration
	BackoffMultiplier float64
}

// DefaultRetryConfig returns FlowGuard's standard retry policy:
// max 3 attempts, 100ms base delay, 2s max delay, 2.0 backoff multiplier,
// full jitter — matching the architecture spec exactly.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:       3,
		BaseDelay:         100 * time.Millisecond,
		MaxDelay:          2 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

// RetryableFunc is a function that can be retried; it must return an error
// that is safe to classify via IsRetryable.
type RetryableFunc func(ctx context.Context) error

// IsRetryable classifies whether an error should trigger a retry, per the
// architecture spec: gRPC codes UNAVAILABLE, RESOURCE_EXHAUSTED, and
// DEADLINE_EXCEEDED are retryable; all others are not.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	switch st.Code() {
	case codes.Unavailable, codes.ResourceExhausted, codes.DeadlineExceeded:
		return true
	default:
		return false
	}
}

// Do executes fn, retrying on retryable errors according to cfg, using full
// jitter backoff. It respects context cancellation between attempts.
func Do(ctx context.Context, cfg RetryConfig, fn RetryableFunc) error {
	var lastErr error
	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := fullJitterDelay(cfg, attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
		lastErr = fn(ctx)
		if lastErr == nil {
			return nil
		}
		if !IsRetryable(lastErr) {
			return lastErr
		}
	}
	return errors.Join(lastErr)
}

func fullJitterDelay(cfg RetryConfig, attempt int) time.Duration {
	exp := cfg.BaseDelay * time.Duration(1<<uint(attempt))
	if exp > cfg.MaxDelay {
		exp = cfg.MaxDelay
	}
	//nolint:gosec // non-cryptographic jitter is acceptable here
	jittered := time.Duration(rand.Int63n(int64(exp) + 1))
	return jittered
}
