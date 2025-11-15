package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// This program generates tls/server.crt and tls/server.key
// The certificate is valid for "localhost" and "server" (for Docker Compose)
func main() {
	certPath := filepath.Join("tls", "server.crt")
	keyPath := filepath.Join("tls", "server.key")

	// Create the tls/ directory if it doesn't exist
	if err := os.MkdirAll("tls", 0755); err != nil {
		log.Fatalf("Failed to create tls directory: %v", err)
	}

	// Generate a private key
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	// Create a certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"grpcstore Dev"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 365), // 1 year
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		BasicConstraintsValid: true,
		// These are the crucial names. "localhost" for local, "server" for Docker.
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:    []string{"localhost", "server"},
	}

	// Create the self-signed certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %v", err)
	}

	// --- Write certificate to tls/server.crt ---
	certOut, err := os.Create(certPath)
	if err != nil {
		log.Fatalf("Failed to open %s for writing: %v", certPath, err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		log.Fatalf("Failed to write data to %s: %v", certPath, err)
	}
	if err := certOut.Close(); err != nil {
		log.Fatalf("Error closing %s: %v", certPath, err)
	}
	log.Printf("Generated %s\n", certPath)

	// --- Write private key to tls/server.key ---
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Failed to open %s for writing: %v", keyPath, err)
	}
	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		log.Fatalf("Unable to marshal private key: %v", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes}); err != nil {
		log.Fatalf("Failed to write data to %s: %v", keyPath, err)
	}
	if err := keyOut.Close(); err != nil {
		log.Fatalf("Error closing %s: %v", keyPath, err)
	}
	log.Printf("Generated %s\n", keyPath)
}
