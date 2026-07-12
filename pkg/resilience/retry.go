package resilience

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RetryInterceptor implements the standard retry envelope from Part 9.2
func RetryInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Define the Exponential Backoff with Full Jitter
		b := backoff.NewExponentialBackOff()
		b.InitialInterval = 100 * time.Millisecond
		b.MaxInterval = 2 * time.Second
		b.Multiplier = 2.0
		// We want a max of 3 attempts, meaning max elapsed time or manually capping retries
		// For simplicity, we wrap it in a WithMaxRetries (3 attempts = 2 retries after initial failure)
		bMax := backoff.WithMaxRetries(b, 2)

		operation := func() error {
			err := invoker(ctx, method, req, reply, cc, opts...)
			if err == nil {
				return nil // Success
			}

			st, ok := status.FromError(err)
			if !ok {
				return err // Non-gRPC error, do not retry
			}

			// Retryable conditions per 9.2
			switch st.Code() {
			case codes.Unavailable, codes.ResourceExhausted, codes.DeadlineExceeded:
				return err // Return error to trigger backoff retry
			case codes.InvalidArgument, codes.NotFound, codes.PermissionDenied, codes.AlreadyExists:
				// Non-retryable
				return backoff.Permanent(err)
			default:
				// Default non-retryable
				return backoff.Permanent(err)
			}
		}

		return backoff.Retry(operation, bMax)
	}
}
