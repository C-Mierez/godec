package gateway

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
)

var (
	entropyBytes   = 32
	plainKeyPrefix = "sk_godec_"
)

func generateKey() (plainKey string, hashedKey string, err error) {
	bytes := make([]byte, entropyBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	// Plaintext API key that will be returned to the tenant one time only
	plainKey = plainKeyPrefix + base64.URLEncoding.EncodeToString(bytes)

	// Create the hash of the API Key
	hash := sha256.Sum256([]byte(plainKey))

	hashedKey = hex.EncodeToString(hash[:])

	return plainKey, hashedKey, nil
}

func validateKey(providedKey string, storedHash string) bool {
	hash := sha256.Sum256([]byte(providedKey))

	hHex := hex.EncodeToString(hash[:])

	return subtle.ConstantTimeCompare([]byte(hHex), []byte(storedHash)) == 1
}
