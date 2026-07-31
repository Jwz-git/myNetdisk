package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"personal-disk/internal/config"
)

func TestSecureHandlerAllowsConfiguredCORSOrigin(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{
			EnableCORS:     true,
			AllowedOrigins: []string{"https://example.com"},
		},
	}
	handler := SecureHandler(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("preflight request should not reach the next handler")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
}

func TestSecureHandlerRateLimitsByClientIP(t *testing.T) {
	resetRateLimitStore()
	cfg := &config.Config{
		Security: config.SecurityConfig{
			EnableRateLimit: true,
			RateLimit: config.RateLimit{
				Requests: 1,
				Window:   60,
			},
		},
	}
	handler := SecureHandler(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.1:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("first status = %d, want %d", rr.Code, http.StatusOK)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.1:1234"
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("second status = %d, want %d", rr.Code, http.StatusTooManyRequests)
	}
}

func resetRateLimitStore() {
	rateLimitStore.Lock()
	defer rateLimitStore.Unlock()
	rateLimitStore.entries = make(map[string]rateLimitEntry)
}
