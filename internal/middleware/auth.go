// Package middleware 提供 HTTP 中间件。
package middleware

import (
	"net/http"
	"strings"

	"supplychain/pkg/httpx"
)

// Auth 基于 Bearer Token 的鉴权中间件。
// token 非空时启用校验，为空则放行所有请求。
func Auth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				httpx.Unauthorized(w, "缺少 Authorization 头")
				return
			}
			if strings.TrimSpace(strings.TrimPrefix(auth, "Bearer ")) != token {
				httpx.Unauthorized(w, "令牌无效")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
