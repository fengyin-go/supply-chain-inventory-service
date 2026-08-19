// Package handler 实现 HTTP 处理器层。
package handler

import (
	"errors"
	"net/http"
	"runtime/debug"
	"time"

	"supplychain/internal/config"
	"supplychain/internal/middleware"
	"supplychain/internal/model"
	"supplychain/internal/service"
	"supplychain/internal/store"
	"supplychain/pkg/httpx"
	"supplychain/pkg/logger"
)

type Server struct {
	svc *service.Service
	log *logger.Logger
	cfg *config.Config
}

func NewServer(svc *service.Service, log *logger.Logger, cfg *config.Config) *Server {
	return &Server{svc: svc, log: log, cfg: cfg}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	s.registerSupplierRoutes(mux)
	s.registerProductRoutes(mux)
	s.registerPurchaseRoutes(mux)
	s.registerInboundRoutes(mux)
	s.registerInspectionRoutes(mux)
	s.registerInventoryRoutes(mux)
	s.registerReportRoutes(mux)
	s.registerImportExportRoutes(mux)
	s.registerReturnRoutes(mux)

	var handler http.Handler = mux
	handler = middleware.Auth(s.authToken())(handler)
	handler = middleware.RateLimit(s.rateLimit())(handler)
	handler = s.loggingMiddleware(s.recoveryMiddleware(handler))
	return handler
}

func (s *Server) authToken() string {
	if s.cfg != nil {
		return s.cfg.AuthToken
	}
	return ""
}

func (s *Server) rateLimit() int {
	if s.cfg != nil {
		return s.cfg.RateLimitPerSec
	}
	return 0
}

func (s *Server) maxPageSize() int {
	if s.cfg != nil && s.cfg.MaxPageSize > 0 {
		return s.cfg.MaxPageSize
	}
	return 100
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.log.Infof("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				s.log.Errorf("panic: %v\n%s", rec, debug.Stack())
				httpx.InternalError(w, "服务器内部错误")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case model.IsValidationError(err):
		httpx.BadRequest(w, err.Error())
	case errors.Is(err, store.ErrNotFound):
		httpx.NotFound(w, err.Error())
	case errors.Is(err, store.ErrConflict):
		httpx.Conflict(w, err.Error())
	default:
		httpx.InternalError(w, err.Error())
	}
}
