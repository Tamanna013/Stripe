package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/flowguard/auth-service/jwt"
)

func main() {
	log.Println("Starting Auth Service...")

	issuer, err := jwt.NewIssuer()
	if err != nil {
		log.Fatalf("Failed to initialize JWT issuer: %v", err)
	}
	_ = issuer

	// Setup minimal HTTP handlers for OIDC and JWKS
	http.HandleFunc("/.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// TODO: Output actual JWKS from the generated RSA public key
		w.Write([]byte(`{"keys": []}`))
	})

	server := &http.Server{Addr: ":8080"}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Auth Service HTTP server failed: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down Auth Service...")
	_ = server.Close()
}
