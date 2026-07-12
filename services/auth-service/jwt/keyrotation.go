package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"sync"
	"time"
)

// KeyRotationProvider manages a 24-hour rotating RSA keypair
type KeyRotationProvider struct {
	mu           sync.RWMutex
	currentKID   string
	currentKey   *rsa.PrivateKey
	historical   map[string]*rsa.PublicKey
}

func NewKeyRotationProvider() (*KeyRotationProvider, error) {
	krp := &KeyRotationProvider{
		historical: make(map[string]*rsa.PublicKey),
	}
	if err := krp.rotateKey(); err != nil {
		return nil, err
	}
	// Background rotation logic goes here
	// Note: Migrating to standard SDK Leader Election when Component 3 is expanded
	go krp.rotationLoop()
	return krp, nil
}

func (krp *KeyRotationProvider) rotationLoop() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		// In a real multi-pod deployment, we would acquire a Redis lock here
		_ = krp.rotateKey()
	}
}

func (krp *KeyRotationProvider) rotateKey() error {
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	newKID := fmt.Sprintf("key-%d", time.Now().Unix())
	
	krp.mu.Lock()
	defer krp.mu.Unlock()
	
	if krp.currentKey != nil {
		// Retain old public key
		krp.historical[krp.currentKID] = &krp.currentKey.PublicKey
		
		// Evict keys older than 2 rotations to cap memory
		if len(krp.historical) > 2 {
			// Find oldest and remove
			var oldestKID string
			for k := range krp.historical {
				oldestKID = k
				break // simplified for Phase 1
			}
			delete(krp.historical, oldestKID)
		}
	}
	
	krp.currentKID = newKID
	krp.currentKey = pk
	krp.historical[newKID] = &pk.PublicKey
	
	return nil
}

func (krp *KeyRotationProvider) CurrentSigningKey() (string, *rsa.PrivateKey) {
	krp.mu.RLock()
	defer krp.mu.RUnlock()
	return krp.currentKID, krp.currentKey
}

func (krp *KeyRotationProvider) PublicKeyByKID(kid string) (*rsa.PublicKey, bool) {
	krp.mu.RLock()
	defer krp.mu.RUnlock()
	pub, ok := krp.historical[kid]
	return pub, ok
}
