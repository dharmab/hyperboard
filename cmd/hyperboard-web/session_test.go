package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"
	"time"
)

func TestSignVerifySession(t *testing.T) {
	secret := "test-secret-key"

	t.Run("valid token", func(t *testing.T) {
		token := signSession(secret)
		if !verifySession(secret, token) {
			t.Error("valid token should verify successfully")
		}
	})

	t.Run("wrong secret rejects", func(t *testing.T) {
		token := signSession(secret)
		if verifySession("wrong-secret", token) {
			t.Error("token signed with different secret should not verify")
		}
	})

	t.Run("tampered token rejects", func(t *testing.T) {
		token := signSession(secret)
		tampered := token + "x"
		if verifySession(secret, tampered) {
			t.Error("tampered token should not verify")
		}
	})

	t.Run("malformed token rejects", func(t *testing.T) {
		malformed := []string{
			"",
			"no-dot-separator",
			"...",
			"not-base64.also-not-base64",
		}
		for _, token := range malformed {
			if verifySession(secret, token) {
				t.Errorf("malformed token %q should not verify", token)
			}
		}
	})

	t.Run("expired token rejects", func(t *testing.T) {
		// Construct a token with an expired timestamp
		expired := time.Now().Add(-(sessionExpiry + time.Hour))
		ts := fmt.Sprintf("%d", expired.Unix())
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(ts))
		sig := mac.Sum(nil)
		token := base64.RawURLEncoding.EncodeToString([]byte(ts)) + "." + base64.RawURLEncoding.EncodeToString(sig)
		if verifySession(secret, token) {
			t.Error("expired token should not verify")
		}
	})
}
