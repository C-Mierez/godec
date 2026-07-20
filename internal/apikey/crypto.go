// Package apikey provides API key generation, validation, and persistence.
package apikey

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
)

var (
	tokenIDEntropyBytes = 16
	secretEntropyBytes  = 32
	tokenIDPrefix       = "gdk_"
	secretPrefix        = "sk_"
)

// generateCredentials generates a split token ID and secret.
// Returns: tokenID (public identifier), secret (private credential), hashedSecret (SHA-256 hash).
func generateCredentials() (tokenID, secret, hashedSecret string, err error) {
	// Generate token ID: gdk_ + base64(16 bytes)
	tokenBytes := make([]byte, tokenIDEntropyBytes)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", "", err
	}
	tokenID = tokenIDPrefix + base64.RawURLEncoding.EncodeToString(tokenBytes)

	// Generate secret: sk_ + base64(32 bytes)
	secretBytes := make([]byte, secretEntropyBytes)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", "", "", err
	}
	secret = secretPrefix + base64.RawURLEncoding.EncodeToString(secretBytes)

	// Hash the secret for storage
	hashedSecret = hashSecret(secret)

	return tokenID, secret, hashedSecret, nil
}

// hashSecret returns the SHA-256 hex digest of the given secret.
func hashSecret(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(hash[:])
}

// validateSecretHash compares a computed hash against the stored hash using
// constant-time comparison to prevent timing attacks.
func validateSecretHash(computedHash, storedHash string) bool {
	return subtle.ConstantTimeCompare([]byte(computedHash), []byte(storedHash)) == 1
}
