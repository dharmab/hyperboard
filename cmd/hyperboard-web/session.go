package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const sessionCookieName = "session"
const sessionExpiry = 30 * 24 * time.Hour

// signSession creates an HMAC-signed session token: base64(timestamp) + "." + base64(hmac)
func signSession(secret string) string {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts))
	sig := mac.Sum(nil)
	return base64.RawURLEncoding.EncodeToString([]byte(ts)) + "." + base64.RawURLEncoding.EncodeToString(sig)
}

// verifySession checks an HMAC-signed session token. Returns true if valid and not expired.
func verifySession(secret, token string) bool {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return false
	}
	tsBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(tsBytes)
	expected := mac.Sum(nil)
	if !hmac.Equal(sigBytes, expected) {
		return false
	}
	// Check expiry
	var ts int64
	if _, err := fmt.Sscanf(string(tsBytes), "%d", &ts); err != nil {
		return false
	}
	return time.Since(time.Unix(ts, 0)) < sessionExpiry
}

func setSessionCookie(w http.ResponseWriter, secret string) {
	token := signSession(secret)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(sessionExpiry.Seconds()),
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}
