package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Issuer struct {
	privateKey *rsa.PrivateKey
}

func NewIssuer() (*Issuer, error) {
	// Generate an RSA keypair for signing JWTs.
	// In production, this would load from KMS or Vault.
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return &Issuer{privateKey: pk}, nil
}

func (i *Issuer) MintToken(userID string, email string, roles []string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"roles": roles,
		"iss":   "flowguard-auth",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(i.privateKey)
}
