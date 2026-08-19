package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"supplychain/pkg/httpx"
)

// RateLimit 基于固定窗口的每 IP 限流中间件。
// perSec 为每秒允许的请求数，<=0 时不限流。
func RateLimit(perSec int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		limiter := newWindowLimiter(perSec)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if perSec <= 0 {
				next.ServeHTTP(w, r)
				return
			}
			ip := clientIP(r)
			if !limiter.allow(ip) {
				httpx.Error(w, http.StatusTooManyRequests, 429, "请求过于频繁，请稍后再试")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// windowLimiter 简单的固定窗口限流器，按 IP 计数。
type windowLimiter struct {
	mu      sync.Mutex
	perSec  int
	windows map[string]*window
}

type window struct {
	start time.Time
	count int
}

func newWindowLimiter(perSec int) *windowLimiter {
	return &windowLimiter{
		perSec:  perSec,
		windows: make(map[string]*window),
	}
}

func (l *windowLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	w, ok := l.windows[key]
	if !ok || now.Sub(w.start) >= time.Second {
		l.windows[key] = &window{start: now, count: 1}
		return true
	}
	if w.count >= l.perSec {
		return false
	}
	w.count++
	return true
}

// clientIP 提取客户端 IP，优先取 X-Forwarded-For 首个地址。
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
