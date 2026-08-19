package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestAuthMiddleware(t *testing.T) {
	// token 为空时放行
	rec := httptest.NewRecorder()
	Auth("")(okHandler()).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("empty token status = %d", rec.Code)
	}

	// token 非空，未携带 → 401
	rec = httptest.NewRecorder()
	Auth("secret")(okHandler()).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing token status = %d", rec.Code)
	}

	// token 错误 → 401
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	rec = httptest.NewRecorder()
	Auth("secret")(okHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("wrong token status = %d", rec.Code)
	}

	// token 正确 → 200
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rec = httptest.NewRecorder()
	Auth("secret")(okHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("correct token status = %d", rec.Code)
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	// perSec <= 0 不限流
	rec := httptest.NewRecorder()
	RateLimit(0)(okHandler()).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("no limit status = %d", rec.Code)
	}

	// perSec = 2：前 2 个放行，第 3 个限流
	handler := RateLimit(2)(okHandler())
	for i := 0; i < 2; i++ {
		rec = httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d status = %d", i, rec.Code)
		}
	}
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("rate limited status = %d, want 429", rec.Code)
	}
}
