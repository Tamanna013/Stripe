package server

import (
	"context"
	
	flowguardv1 "github.com/flowguard/protos/gen/go/flowguard/v1"
	"github.com/flowguard/auth-service/jwt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	flowguardv1.UnimplementedAuthServer
	issuer *jwt.Issuer
}

func NewServer(issuer *jwt.Issuer) *Server {
	return &Server{issuer: issuer}
}

func (s *Server) Login(ctx context.Context, req *flowguardv1.LoginRequest) (*flowguardv1.LoginResponse, error) {
	// 1. Validate OIDC code (mocked for phase 1)
	if len(req.OidcCode) > 2048 {
		return nil, status.Error(codes.InvalidArgument, "oidc code too long")
	}

	// 2. Fetch User from DB (mocked for phase 1)
	subject := "sub-123"
	email := "engineer@flowguard.internal"
	roles := []string{"sre-admin"}

	// 3. Issue Token
	token, _, err := s.issuer.IssueAccessToken(subject, email, roles)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to issue token")
	}

	return &flowguardv1.LoginResponse{
		AccessToken: token,
		RefreshToken: "mock_refresh_token_xyz", // would be hashed in DB
		ExpiresInSeconds: 900, // 15 mins
	}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *flowguardv1.ValidateTokenRequest) (*flowguardv1.ValidateTokenResponse, error) {
	claims, err := s.issuer.ValidateAccessToken(req.AccessToken)
	if err != nil {
		return &flowguardv1.ValidateTokenResponse{
			Valid: false,
			InvalidReason: err.Error(),
		}, nil
	}

	return &flowguardv1.ValidateTokenResponse{
		Valid: true,
		Subject: claims.Subject,
		Email: claims.Email,
		Roles: claims.Roles,
	}, nil
}

func (s *Server) GetJWKS(ctx context.Context, req *flowguardv1.GetJWKSRequest) (*flowguardv1.GetJWKSResponse, error) {
	// Serializes current + historical keys to JWKS JSON (omitted explicit string build for phase 1 outline)
	return &flowguardv1.GetJWKSResponse{
		JwksJson: `{"keys":[]}`,
	}, nil
}

func (s *Server) IssueServiceCert(ctx context.Context, req *flowguardv1.IssueServiceCertRequest) (*flowguardv1.ServiceCert, error) {
	return nil, status.Error(codes.Unimplemented, "SPIFFE/SPIRE integration deferred to a later phase")
}
