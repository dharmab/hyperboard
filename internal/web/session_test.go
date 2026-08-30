package web

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSignVerifySession(t *testing.T) {
	t.Parallel()
	secret := "test-secret-key"

	t.Run("valid token", func(t *testing.T) {
		t.Parallel()
		token := signSession(secret)
		verified := verifySession(secret, token)
		assert.True(t, verified, "valid token should verify successfully")
	})

	t.Run("wrong secret rejects", func(t *testing.T) {
		t.Parallel()
		token := signSession(secret)
		verified := verifySession("wrong-secret", token)
		assert.False(t, verified, "token signed with different secret should not verify")
	})

	t.Run("tampered token rejects", func(t *testing.T) {
		t.Parallel()
		token := signSession(secret)
		tampered := token + "x"
		verified := verifySession(secret, tampered)
		assert.False(t, verified, "tampered token should not verify")
	})

	t.Run("malformed token rejects", func(t *testing.T) {
		t.Parallel()
		malformed := []string{
			"",
			"no-dot-separator",
			"...",
			"not-base64.also-not-base64",
		}
		for _, token := range malformed {
			verified := verifySession(secret, token)
			assert.False(t, verified, "malformed token %q should not verify", token)
		}
	})

	t.Run("expired token rejects", func(t *testing.T) {
		t.Parallel()
		// Construct a token with an expired timestamp
		expired := time.Now().Add(-(sessionExpiry + time.Hour))
		ts := strconv.FormatInt(expired.Unix(), 10)
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(ts))
		sig := mac.Sum(nil)
		token := base64.RawURLEncoding.EncodeToString([]byte(ts)) + "." + base64.RawURLEncoding.EncodeToString(sig)
		verified := verifySession(secret, token)
		assert.False(t, verified, "expired token should not verify")
	})
}
