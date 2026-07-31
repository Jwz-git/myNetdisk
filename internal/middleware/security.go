package middleware

import (
	"net"
	"net/http"
	"personal-disk/internal/config"
	"strings"
	"sync"
	"time"
)

type rateLimitEntry struct {
	count   int
	resetAt time.Time
}

var rateLimitStore = struct {
	sync.Mutex
	entries map[string]rateLimitEntry
}{
	entries: make(map[string]rateLimitEntry),
}

// SecureHandler 为应用入口统一应用 CORS 和限流策略。
func SecureHandler(cfg *config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cfg.Security.EnableCORS && applyCORS(w, r, cfg) {
			return
		}
		if cfg.Security.EnableRateLimit && !allowRequest(r, cfg) {
			http.Error(w, "请求过于频繁", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func applyCORS(w http.ResponseWriter, r *http.Request, cfg *config.Config) bool {
	origin := r.Header.Get("Origin")
	if origin == "" || !isAllowedOrigin(origin, cfg.Security.AllowedOrigins) {
		return r.Method == http.MethodOptions
	}

	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Vary", "Origin")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if origin == allowedOrigin {
			return true
		}
	}
	return false
}

func allowRequest(r *http.Request, cfg *config.Config) bool {
	if cfg.Security.RateLimit.Requests <= 0 || cfg.Security.RateLimit.Window <= 0 {
		return true
	}

	key := clientIP(r)
	now := time.Now()
	window := time.Duration(cfg.Security.RateLimit.Window) * time.Second

	rateLimitStore.Lock()
	defer rateLimitStore.Unlock()

	entry := rateLimitStore.entries[key]
	if now.After(entry.resetAt) {
		rateLimitStore.entries[key] = rateLimitEntry{
			count:   1,
			resetAt: now.Add(window),
		}
		return true
	}

	if entry.count >= cfg.Security.RateLimit.Requests {
		return false
	}

	entry.count++
	rateLimitStore.entries[key] = entry
	return true
}

func clientIP(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
