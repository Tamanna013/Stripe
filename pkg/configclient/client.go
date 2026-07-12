// Package configclient provides a client for FlowGuard's Config Service,
// with local in-memory caching and last-known-good fallback on disconnect.
package configclient

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	flowguardv1 "github.com/flowguard/protos/gen/go/flowguard/v1"
)

// Client wraps the generated gRPC Config client with caching and fallback.
type Client struct {
	grpc  flowguardv1.ConfigClient
	mu    sync.RWMutex
	cache map[string]string // key -> value_json
}

// New creates a Client wrapping the given generated gRPC client.
func New(grpcClient flowguardv1.ConfigClient) *Client {
	return &Client{
		grpc:  grpcClient,
		cache: make(map[string]string),
	}
}

// GetJSON fetches the config value for key, unmarshalling into dest.
// On RPC failure, it falls back to the last cached value for key if present;
// if no cached value exists, it returns the underlying error.
func (c *Client) GetJSON(ctx context.Context, key string, dest any) error {
	resp, err := c.grpc.GetConfig(ctx, &flowguardv1.GetConfigRequest{Key: key})
	if err != nil {
		c.mu.RLock()
		cached, ok := c.cache[key]
		c.mu.RUnlock()
		if !ok {
			return fmt.Errorf("fetching config key %q (no cached fallback available): %w", key, err)
		}
		return json.Unmarshal([]byte(cached), dest)
	}
	c.mu.Lock()
	c.cache[key] = resp.ValueJson
	c.mu.Unlock()
	return json.Unmarshal([]byte(resp.ValueJson), dest)
}

// Watch starts a background goroutine watching keys matching keyPrefix,
// updating the local cache as new values arrive. It returns immediately;
// the watch runs until ctx is cancelled.
func (c *Client) Watch(ctx context.Context, keyPrefix string) error {
	stream, err := c.grpc.WatchConfig(ctx, &flowguardv1.WatchConfigRequest{KeyPrefix: keyPrefix})
	if err != nil {
		return fmt.Errorf("starting config watch for prefix %q: %w", keyPrefix, err)
	}
	go func() {
		for {
			val, err := stream.Recv()
			if err != nil {
				return // stream closed; caller's cached values remain as last-known-good
			}
			c.mu.Lock()
			c.cache[val.Key] = val.ValueJson
			c.mu.Unlock()
		}
	}()
	return nil
}
