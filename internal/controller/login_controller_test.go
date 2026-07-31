package controller

import (
	"strings"
	"testing"

	"personal-disk/internal/config"
)

func testSessionConfig() *config.Config {
	return &config.Config{
		Admin: config.AdminConfig{
			Username: "admin",
			Password: "strong-password",
		},
		Session: config.SessionConfig{
			CookieName:  "admin_session",
			CookieValue: "session-secret",
			CookiePath:  "/",
			MaxAge:      3600,
		},
	}
}

func TestSessionTokenVerification(t *testing.T) {
	cfg := testSessionConfig()

	token, err := newSessionToken(cfg)
	if err != nil {
		t.Fatalf("newSessionToken returned error: %v", err)
	}
	if !verifySessionToken(cfg, token) {
		t.Fatal("expected a freshly created token to verify")
	}
}

func TestSessionTokenRejectsTampering(t *testing.T) {
	cfg := testSessionConfig()

	token, err := newSessionToken(cfg)
	if err != nil {
		t.Fatalf("newSessionToken returned error: %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		t.Fatalf("unexpected token shape: %q", token)
	}
	tamperedToken := parts[0] + ".tampered"

	if verifySessionToken(cfg, tamperedToken) {
		t.Fatal("expected tampered token to be rejected")
	}
}

func TestSessionTokenRejectsExpiredToken(t *testing.T) {
	cfg := testSessionConfig()
	cfg.Session.MaxAge = -1

	token, err := newSessionToken(cfg)
	if err != nil {
		t.Fatalf("newSessionToken returned error: %v", err)
	}
	if verifySessionToken(cfg, token) {
		t.Fatal("expected expired token to be rejected")
	}
}
