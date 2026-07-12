package resilience

import (
	"time"

	"github.com/sony/gobreaker"
)

// NewStandardCircuitBreaker returns a circuit breaker configured to
// FlowGuard's standard policy from the architecture spec.
func NewStandardCircuitBreaker(name string) *gobreaker.CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     15 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if counts.Requests < 10 {
				return false
			}
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return failureRatio >= 0.5
		},
		IsSuccessful: func(err error) bool {
			return err == nil
		},
	}
	return gobreaker.NewCircuitBreaker(settings)
}

// NewCircuitBreakerWithThreshold returns a circuit breaker with a custom
// failure ratio threshold, for dependency-specific tuning (e.g., stricter
// for a primary database, looser for a best-effort notifier).
func NewCircuitBreakerWithThreshold(name string, failureRatioThreshold float64, minRequests uint32) *gobreaker.CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     15 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if counts.Requests < minRequests {
				return false
			}
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return failureRatio >= failureRatioThreshold
		},
		IsSuccessful: func(err error) bool {
			return err == nil
		},
	}
	return gobreaker.NewCircuitBreaker(settings)
}
