package apikey

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

func hashKey(plainKey string) string {
	hash := sha256.Sum256([]byte(plainKey))
	return hex.EncodeToString(hash[:])
}

func generateKey() (plainKey string, hashedKey string, err error) {
	bytes := make([]byte, entropyBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	plainKey = plainKeyPrefix + base64.URLEncoding.EncodeToString(bytes)
	hashedKey = hashKey(plainKey)

	return plainKey, hashedKey, nil
}

func validatePlainKey(plainKey string, storedHash string) bool {
	hHex := hashKey(plainKey)

	return subtle.ConstantTimeCompare([]byte(hHex), []byte(storedHash)) == 1
}

func validateHashedKey(providedHash string, storedHash string) bool {
	return subtle.ConstantTimeCompare([]byte(providedHash), []byte(storedHash)) == 1
}
