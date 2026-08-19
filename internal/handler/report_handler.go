package handler

import (
	"net/http"
	"strconv"

	"supplychain/pkg/httpx"
)

func (s *Server) registerReportRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/reports/stock", s.stockReport)
	mux.HandleFunc("GET /api/reports/quality", s.qualityStats)
	mux.HandleFunc("GET /api/reports/supplier-ranking", s.supplierRanking)
	mux.HandleFunc("GET /api/reports/inventory-value", s.inventoryValue)
	mux.HandleFunc("GET /api/reports/purchase-summary", s.purchaseSummary)
	mux.HandleFunc("GET /api/reports/movements", s.movementStats)
	mux.HandleFunc("GET /api/reports/categories", s.categoryStats)
}

func (s *Server) stockReport(w http.ResponseWriter, r *http.Request) {
	onlyLow := r.URL.Query().Get("low") == "true"
	report, err := s.svc.StockReport(onlyLow)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, report)
}

func (s *Server) qualityStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.svc.QualityStats()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, stats)
}

func (s *Server) supplierRanking(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	ranking, err := s.svc.SupplierRanking(limit)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, ranking)
}

func (s *Server) inventoryValue(w http.ResponseWriter, r *http.Request) {
	value, err := s.svc.InventoryValue()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, value)
}

func (s *Server) purchaseSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := s.svc.PurchaseSummary()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, summary)
}

func (s *Server) movementStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.svc.MovementStats()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, stats)
}

func (s *Server) categoryStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.svc.CategoryStats()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, stats)
}
