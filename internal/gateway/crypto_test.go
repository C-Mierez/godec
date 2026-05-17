package gateway

import (
	"encoding/base64"
	"encoding/hex"
	"regexp"
	"strings"
	"testing"
)

// TestGenerateKeyReturnsNonEmpty verifies that generateKey returns non-empty values
func TestGenerateKeyReturnsNonEmpty(t *testing.T) {
	plainKey, hashedKey, err := generateKey()

	if err != nil {
		t.Fatalf("generateKey() returned error: %v", err)
	}

	if plainKey == "" {
		t.Error("plainKey is empty")
	}

	if hashedKey == "" {
		t.Error("hashedKey is empty")
	}
}

// TestGenerateKeyPlainKeyFormat verifies the plainKey has the correct format
func TestGenerateKeyPlainKeyFormat(t *testing.T) {
	plainKey, _, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() returned error: %v", err)
	}

	// Check prefix
	if !strings.HasPrefix(plainKey, plainKeyPrefix) {
		t.Errorf("plainKey does not start with prefix %q, got: %q", plainKeyPrefix, plainKey[:len(plainKeyPrefix)])
	}

	// Extract the encoded part
	encodedPart := strings.TrimPrefix(plainKey, plainKeyPrefix)

	// Verify it's valid base64
	decoded, err := base64.URLEncoding.DecodeString(encodedPart)
	if err != nil {
		t.Errorf("plainKey encoded part is not valid base64: %v", err)
	}

	// Verify decoded length matches entropyBytes
	if len(decoded) != entropyBytes {
		t.Errorf("decoded key length mismatch: expected %d bytes, got %d", entropyBytes, len(decoded))
	}
}

// TestGenerateKeyPlainKeyLength verifies plainKey has expected length
func TestGenerateKeyPlainKeyLength(t *testing.T) {
	plainKey, _, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() returned error: %v", err)
	}

	// Base64 encoding of 32 bytes produces: ceil(32/3)*4 = 44 characters
	// Plus prefix length (9 for "sk_godec_")
	expectedLength := len(plainKeyPrefix) + 44
	if len(plainKey) != expectedLength {
		t.Errorf("plainKey length mismatch: expected %d, got %d. Value: %q", expectedLength, len(plainKey), plainKey)
	}
}

// TestGenerateKeyHashedKeyFormat verifies hashedKey is valid hex and correct length
func TestGenerateKeyHashedKeyFormat(t *testing.T) {
	_, hashedKey, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() returned error: %v", err)
	}

	// SHA-256 produces 32 bytes, hex encoding produces 64 characters
	if len(hashedKey) != 64 {
		t.Errorf("hashedKey length mismatch: expected 64, got %d", len(hashedKey))
	}

	// Verify it's valid hex (lowercase only)
	decoded, err := hex.DecodeString(hashedKey)
	if err != nil {
		t.Errorf("hashedKey is not valid hex: %v", err)
	}

	if len(decoded) != 32 {
		t.Errorf("decoded hash length mismatch: expected 32 bytes, got %d", len(decoded))
	}

	// Verify it's lowercase
	if hashedKey != strings.ToLower(hashedKey) {
		t.Errorf("hashedKey contains uppercase characters, should be lowercase: %q", hashedKey)
	}
}

// TestGenerateKeyRandomness verifies that multiple calls produce different keys
func TestGenerateKeyRandomness(t *testing.T) {
	keys := make(map[string]bool)
	hashes := make(map[string]bool)

	// Generate 10 keys and verify no duplicates
	for i := 0; i < 10; i++ {
		plainKey, hashedKey, err := generateKey()
		if err != nil {
			t.Fatalf("generateKey() returned error on iteration %d: %v", i, err)
		}

		if keys[plainKey] {
			t.Errorf("duplicate plainKey generated on iteration %d: %q", i, plainKey)
		}
		keys[plainKey] = true

		if hashes[hashedKey] {
			t.Errorf("duplicate hashedKey generated on iteration %d: %q", i, hashedKey)
		}
		hashes[hashedKey] = true
	}

	if len(keys) != 10 {
		t.Errorf("expected 10 unique plainKeys, got %d", len(keys))
	}

	if len(hashes) != 10 {
		t.Errorf("expected 10 unique hashedKeys, got %d", len(hashes))
	}
}

// TestGenerateKeyPlainAndHashAreUnrelated verifies plainKey and hashedKey are sufficiently different
func TestGenerateKeyPlainAndHashAreUnrelated(t *testing.T) {
	plainKey, hashedKey, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() returned error: %v", err)
	}

	// They should not be equal (obviously)
	if plainKey == hashedKey {
		t.Error("plainKey and hashedKey are the same")
	}

	// plainKey should not contain the hash
	if strings.Contains(plainKey, hashedKey) {
		t.Error("plainKey contains hashedKey")
	}

	// hashedKey should not contain plainKey
	if strings.Contains(hashedKey, plainKey) {
		t.Error("hashedKey contains plainKey")
	}
}

// TestValidateKeyCorrectKey verifies validateKey returns true for the correct key
func TestValidateKeyCorrectKey(t *testing.T) {
	plainKey, hashedKey, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() returned error: %v", err)
	}

	if !validateKey(plainKey, hashedKey) {
		t.Errorf("validateKey returned false for correct key. plainKey: %q, hashedKey: %q", plainKey, hashedKey)
	}
}

// TestValidateKeyIncorrectKey verifies validateKey returns false for wrong keys
func TestValidateKeyIncorrectKey(t *testing.T) {
	_, hashedKey, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() returned error: %v", err)
	}

	wrongKey := "sk_godec_wrongkeythatdoesnotmatch"
	if validateKey(wrongKey, hashedKey) {
		t.Error("validateKey returned true for wrong key")
	}
}

// TestValidateKeyEmptyPlainKey verifies validateKey returns false for empty plainKey
func TestValidateKeyEmptyPlainKey(t *testing.T) {
	_, hashedKey, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() returned error: %v", err)
	}

	if validateKey("", hashedKey) {
		t.Error("validateKey returned true for empty plainKey")
	}
}

// TestValidateKeyEmptyHash verifies validateKey returns false for empty hash
func TestValidateKeyEmptyHash(t *testing.T) {
	plainKey, _, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() returned error: %v", err)
	}

	if validateKey(plainKey, "") {
		t.Error("validateKey returned true for empty hash")
	}
}

// TestValidateKeyBothEmpty verifies validateKey returns false when both are empty
func TestValidateKeyBothEmpty(t *testing.T) {
	if validateKey("", "") {
		t.Error("validateKey returned true when both plainKey and hash are empty")
	}
}

// TestValidateKeyMalformedHash verifies validateKey returns false for invalid hex hash
func TestValidateKeyMalformedHash(t *testing.T) {
	plainKey, _, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() returned error: %v", err)
	}

	malformedHash := "not-a-valid-hex-string-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	if validateKey(plainKey, malformedHash) {
		t.Error("validateKey returned true for malformed hash")
	}
}

// TestValidateKeyOffByOneCharacterInHash verifies false for single character difference
func TestValidateKeyOffByOneCharacterInHash(t *testing.T) {
	plainKey, hashedKey, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() returned error: %v", err)
	}

	// Change the last character of the hash
	modifiedHash := hashedKey[:len(hashedKey)-1] + "0"
	if hashedKey[len(hashedKey)-1:] == "0" {
		modifiedHash = hashedKey[:len(hashedKey)-1] + "1"
	}

	if validateKey(plainKey, modifiedHash) {
		t.Errorf("validateKey returned true for modified hash. Original: %q, Modified: %q", hashedKey, modifiedHash)
	}
}

// TestValidateKeyOffByOneCharacterInPlainKey verifies false for single character difference
func TestValidateKeyOffByOneCharacterInPlainKey(t *testing.T) {
	plainKey, hashedKey, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() returned error: %v", err)
	}

	// Change the last character of the plainKey
	modifiedKey := plainKey[:len(plainKey)-1] + "X"
	if plainKey[len(plainKey)-1:] == "X" {
		modifiedKey = plainKey[:len(plainKey)-1] + "Y"
	}

	if validateKey(modifiedKey, hashedKey) {
		t.Errorf("validateKey returned true for modified plainKey. Original: %q, Modified: %q", plainKey, modifiedKey)
	}
}

// TestValidateKeyUnicodeHandling verifies validateKey handles unicode properly
func TestValidateKeyUnicodeHandling(t *testing.T) {
	_, hashedKey, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() returned error: %v", err)
	}

	unicodeKey := "sk_godec_こんにちは世界"
	if validateKey(unicodeKey, hashedKey) {
		t.Error("validateKey returned true for unicode key")
	}
}

// TestValidateKeyWhitespaceHandling verifies validateKey handles whitespace properly
func TestValidateKeyWhitespaceHandling(t *testing.T) {
	plainKey, hashedKey, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() returned error: %v", err)
	}

	// Keys with extra spaces should not match
	keyWithSpaces := plainKey + " "
	if validateKey(keyWithSpaces, hashedKey) {
		t.Error("validateKey returned true for plainKey with trailing space")
	}

	keyWithLeadingSpace := " " + plainKey
	if validateKey(keyWithLeadingSpace, hashedKey) {
		t.Error("validateKey returned true for plainKey with leading space")
	}
}

// TestValidateKeyDeterministic verifies that the same plainKey always produces the same hash
func TestValidateKeyDeterministic(t *testing.T) {
	_ = "sk_godec_testkey123"

	// Validate multiple times - should always return same result
	results := make([]bool, 5)
	_ = ""

	// First call to get the hash
	plainKeyForGeneration, generatedHash, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() returned error: %v", err)
	}

	// Verify the correct key validates
	if !validateKey(plainKeyForGeneration, generatedHash) {
		t.Error("first validation of correct key failed")
	}

	// Verify multiple validations of same key return same result
	for i := 0; i < 5; i++ {
		results[i] = validateKey(plainKeyForGeneration, generatedHash)
	}

	for i := 0; i < 5; i++ {
		if !results[i] {
			t.Errorf("validation result %d is false, expected true", i)
		}
	}
}

// TestValidateKeyCaseSensitivity verifies that hash comparison is case-sensitive
func TestValidateKeyCaseSensitivity(t *testing.T) {
	plainKey, hashedKey, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() returned error: %v", err)
	}

	// Hashedkey is lowercase, try uppercase version
	uppercaseHash := strings.ToUpper(hashedKey)

	// This should fail because generateKey returns lowercase hex
	if validateKey(plainKey, uppercaseHash) {
		t.Error("validateKey returned true for uppercase hash (constant-time compare should fail)")
	}
}

// TestValidateKeyParseableButWrongHash verifies false for hex that decodes but is wrong
func TestValidateKeyParseableButWrongHash(t *testing.T) {
	plainKey, hashedKey, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() returned error: %v", err)
	}

	// Create a valid hex string of correct length but different content
	wrongHashButValidHex := strings.Repeat("a", 64)

	// Make sure it's actually different
	if wrongHashButValidHex == hashedKey {
		wrongHashButValidHex = strings.Repeat("b", 64)
	}

	if validateKey(plainKey, wrongHashButValidHex) {
		t.Errorf("validateKey returned true for wrong but valid hex hash. Expected: %q, Got: %q", hashedKey, wrongHashButValidHex)
	}
}

// TestValidateKeyHashTooShort verifies false for hash that's too short
func TestValidateKeyHashTooShort(t *testing.T) {
	plainKey, _, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() returned error: %v", err)
	}

	shortHash := "abcdef0123456789"
	if validateKey(plainKey, shortHash) {
		t.Error("validateKey returned true for hash that's too short")
	}
}

// TestValidateKeyHashTooLong verifies false for hash that's too long
func TestValidateKeyHashTooLong(t *testing.T) {
	plainKey, _, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() returned error: %v", err)
	}

	longHash := strings.Repeat("a", 128)
	if validateKey(plainKey, longHash) {
		t.Error("validateKey returned true for hash that's too long")
	}
}

// TestValidateKeySpecialCharactersInPlainKey verifies special characters don't accidentally match
func TestValidateKeySpecialCharactersInPlainKey(t *testing.T) {
	_, hashedKey, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() returned error: %v", err)
	}

	specialKeys := []string{
		"sk_godec_!@#$%^&*()",
		"sk_godec_\n\t\r",
		"sk_godec_<script>alert('xss')</script>",
		"sk_godec_'; DROP TABLE keys; --",
	}

	for _, key := range specialKeys {
		if validateKey(key, hashedKey) {
			t.Errorf("validateKey returned true for special character key: %q", key)
		}
	}
}

// BenchmarkGenerateKey benchmarks the key generation function
func BenchmarkGenerateKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, err := generateKey()
		if err != nil {
			b.Fatalf("generateKey() returned error: %v", err)
		}
	}
}

// BenchmarkValidateKey benchmarks the key validation function
func BenchmarkValidateKey(b *testing.B) {
	plainKey, hashedKey, err := generateKey()
	if err != nil {
		b.Fatalf("generateKey() returned error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateKey(plainKey, hashedKey)
	}
}

// TestGenerateKeyPrefixExactMatch verifies the exact prefix is used
func TestGenerateKeyPrefixExactMatch(t *testing.T) {
	plainKey, _, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() returned error: %v", err)
	}

	expectedPrefix := "sk_godec_"
	if !strings.HasPrefix(plainKey, expectedPrefix) {
		t.Errorf("plainKey prefix mismatch: expected %q, got %q", expectedPrefix, plainKey[:len(expectedPrefix)])
	}

	if len(plainKey) <= len(expectedPrefix) {
		t.Errorf("plainKey is not longer than prefix: %q", plainKey)
	}
}

// TestNoLeakedInformationBetweenGenerations verifies keys don't leak information between calls
func TestNoLeakedInformationBetweenGenerations(t *testing.T) {
	// Generate multiple keys rapidly
	type keyPair struct {
		plain  string
		hashed string
	}

	keys := make([]keyPair, 100)
	for i := 0; i < 100; i++ {
		plain, hashed, err := generateKey()
		if err != nil {
			t.Fatalf("generateKey() at iteration %d returned error: %v", i, err)
		}
		keys[i] = keyPair{plain, hashed}
	}

	// Verify no hashed key from one generation validates against a different key
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if validateKey(keys[i].plain, keys[j].hashed) {
				t.Errorf("key %d validated against hash from key %d (should not happen)", i, j)
			}
			if validateKey(keys[j].plain, keys[i].hashed) {
				t.Errorf("key %d validated against hash from key %d (should not happen)", j, i)
			}
		}
	}
}

// TestPlainKeyCharacterSet verifies plainKey only contains expected characters
func TestPlainKeyCharacterSet(t *testing.T) {
	plainKey, _, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() returned error: %v", err)
	}

	// URL-safe base64 uses: A-Z, a-z, 0-9, -, _, =
	// Plus the prefix which contains: a-z, _
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_\-=]+$`)
	if !validPattern.MatchString(plainKey) {
		t.Errorf("plainKey contains unexpected characters: %q", plainKey)
	}
}

// TestHashedKeyCharacterSet verifies hashedKey only contains hex characters
func TestHashedKeyCharacterSet(t *testing.T) {
	_, hashedKey, err := generateKey()
	if err != nil {
		t.Fatalf("generateKey() returned error: %v", err)
	}

	// Hex encoding should only contain 0-9, a-f (lowercase)
	validPattern := regexp.MustCompile(`^[0-9a-f]+$`)
	if !validPattern.MatchString(hashedKey) {
		t.Errorf("hashedKey contains non-hex characters: %q", hashedKey)
	}
}
