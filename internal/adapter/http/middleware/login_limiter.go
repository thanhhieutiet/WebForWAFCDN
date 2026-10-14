package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"labshop/internal/adapter/http/presenter"
	"labshop/internal/domain/errs"
)

type ipRecord struct {
	failedAttempts int
	lockedUntil    time.Time
	lastAttempt    time.Time
}

// LoginLimiter provides in-memory rate limiting and brute-force protection for login endpoints.
type LoginLimiter struct {
	mu              sync.Mutex
	records         map[string]*ipRecord
	maxFailures     int
	lockoutDuration time.Duration
	windowDuration  time.Duration
}

// Global default login limiter
var DefaultLoginLimiter = NewLoginLimiter(5, 15*time.Minute, 5*time.Minute)

func NewLoginLimiter(maxFailures int, lockoutDuration, windowDuration time.Duration) *LoginLimiter {
	l := &LoginLimiter{
		records:         make(map[string]*ipRecord),
		maxFailures:     maxFailures,
		lockoutDuration: lockoutDuration,
		windowDuration:  windowDuration,
	}

	// Periodically clean up stale records every 10 minutes
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		for range ticker.C {
			l.mu.Lock()
			now := time.Now()
			for ip, rec := range l.records {
				if now.After(rec.lockedUntil) && now.Sub(rec.lastAttempt) > l.windowDuration {
					delete(l.records, ip)
				}
			}
			l.mu.Unlock()
		}
	}()

	return l
}

// GetClientIP extracts the client IP address considering proxy headers.
func GetClientIP(r *http.Request) string {
	// If X-Forwarded-For is set by trusted reverse proxy (Nginx / WAAP), use first IP
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

// CheckAllowed checks whether an IP is currently locked out.
func (l *LoginLimiter) CheckAllowed(ip string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	rec, exists := l.records[ip]
	if !exists {
		return true, 0
	}

	now := time.Now()
	if now.Before(rec.lockedUntil) {
		return false, rec.lockedUntil.Sub(now)
	}

	// If lockout expired, reset
	if !rec.lockedUntil.IsZero() && now.After(rec.lockedUntil) {
		delete(l.records, ip)
		return true, 0
	}

	return true, 0
}

// RecordFailure records a failed login attempt. Returns true if IP is now locked out.
func (l *LoginLimiter) RecordFailure(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	rec, exists := l.records[ip]
	if !exists || now.Sub(rec.lastAttempt) > l.windowDuration {
		rec = &ipRecord{failedAttempts: 0}
		l.records[ip] = rec
	}

	rec.failedAttempts++
	rec.lastAttempt = now

	if rec.failedAttempts >= l.maxFailures {
		rec.lockedUntil = now.Add(l.lockoutDuration)
		return true
	}

	return false
}

// RecordSuccess clears any failure records for an IP upon successful login.
func (l *LoginLimiter) RecordSuccess(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.records, ip)
}

// Middleware returns an HTTP handler middleware to protect the login endpoint.
func (l *LoginLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := GetClientIP(r)
		allowed, remaining := l.CheckAllowed(ip)
		if !allowed {
			w.Header().Set("Retry-After", "900")
			presenter.Error(w, r, errs.NewBusinessRule("too_many_attempts",
				"Tài khoản hoặc IP đã bị tạm khóa do nhập sai mật khẩu quá 5 lần. Vui lòng thử lại sau "+remaining.Round(time.Second).String()+"."))
			return
		}
		next.ServeHTTP(w, r)
	})
}
