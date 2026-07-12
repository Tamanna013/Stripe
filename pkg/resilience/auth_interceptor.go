package resilience

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthInterceptor validates the FlowGuard JWT on incoming requests
func AuthInterceptor(jwksURL string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// Mocked validation for Phase 1
		// In production, this would fetch from jwksURL (caching the keys) and validate the JWT signature.
		
		// If validation fails:
		// return nil, status.Error(codes.Unauthenticated, "invalid or missing JWT")

		// If successful, we can inject claims into the context and proceed:
		return handler(ctx, req)
	}
}
