package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

const clientCookieName = "datastar-client-id"

func getOrSetClientID(w http.ResponseWriter, r *http.Request) string {
	if cookie, err := r.Cookie(clientCookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	id := newClientID()
	http.SetCookie(w, &http.Cookie{
		Name:     clientCookieName,
		Value:    id,
		Path:     "/",
		Expires:  time.Now().Add(1 * time.Hour),
		MaxAge:   3600,
		SameSite: http.SameSiteLaxMode,
	})
	return id
}

func newClientID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
