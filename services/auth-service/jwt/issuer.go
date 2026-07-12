// Package jwt implements FlowGuard's session token issuance and validation,
// using RS256 with automatic key rotation (see keyrotation.go).
package jwt

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
	issuer          = "flowguard-auth"
)

// Claims represents the FlowGuard access token claims.
type Claims struct {
	jwt.RegisteredClaims
	Email string   `json:"email"`
	Roles []string `json:"roles"`
}

// Issuer issues and validates FlowGuard session tokens.
type Issuer struct {
	keyProvider KeyProvider
}

// KeyProvider supplies the current signing key and historical keys for
// validation across rotation boundaries (see keyrotation.go for the
// concrete implementation backing this interface).
type KeyProvider interface {
	CurrentSigningKey() (kid string, key *rsa.PrivateKey)
	PublicKeyByKID(kid string) (*rsa.PublicKey, bool)
}

// NewIssuer creates an Issuer backed by the given KeyProvider.
func NewIssuer(kp KeyProvider) *Issuer {
	return &Issuer{keyProvider: kp}
}

// IssueAccessToken issues a 15-minute access token for the given subject.
func (i *Issuer) IssueAccessToken(subject, email string, roles []string) (string, time.Time, error) {
	kid, privKey := i.keyProvider.CurrentSigningKey()
	expiresAt := time.Now().Add(accessTokenTTL)
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		Email: email,
		Roles: roles,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	signed, err := token.SignedString(privKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("signing access token for subject %q: %w", subject, err)
	}
	return signed, expiresAt, nil
}

// ValidateAccessToken parses and validates an access token, returning its claims.
func (i *Issuer) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		kid, ok := t.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("token missing kid header")
		}
		pub, found := i.keyProvider.PublicKeyByKID(kid)
		if !found {
			return nil, fmt.Errorf("unknown signing key id %q", kid)
		}
		return pub, nil
	}, jwt.WithValidMethods([]string{"RS256"}), jwt.WithIssuer(issuer))
	if err != nil {
		return nil, fmt.Errorf("validating access token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("token failed validation")
	}
	return claims, nil
}
