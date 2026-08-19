package handler

import (
	"net/http"

	"supplychain/internal/model"
	"supplychain/pkg/httpx"
)

func (s *Server) registerInventoryRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/inbound-orders/{id}/stock", s.stockInbound)
	mux.HandleFunc("POST /api/stock/outbound", s.outboundStock)
	mux.HandleFunc("POST /api/stock/adjust", s.adjustStock)
	mux.HandleFunc("GET /api/stock/{product_id}", s.getStock)
	mux.HandleFunc("GET /api/batches", s.listBatches)
	mux.HandleFunc("GET /api/movements", s.listMovements)
}

func (s *Server) stockInbound(w http.ResponseWriter, r *http.Request) {
	inbound, err := s.svc.StockInbound(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, inbound)
}

type outboundStockRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Note      string `json:"note"`
}

func (s *Server) outboundStock(w http.ResponseWriter, r *http.Request) {
	var req outboundStockRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	if err := s.svc.OutboundStock(req.ProductID, req.Quantity, req.Note); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, map[string]string{"status": "ok"})
}

type adjustStockRequest struct {
	ProductID string `json:"product_id"`
	Delta     int    `json:"delta"`
	Note      string `json:"note"`
}

func (s *Server) adjustStock(w http.ResponseWriter, r *http.Request) {
	var req adjustStockRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	if err := s.svc.AdjustStock(req.ProductID, req.Delta, req.Note); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, map[string]string{"status": "ok"})
}

func (s *Server) getStock(w http.ResponseWriter, r *http.Request) {
	stock, err := s.svc.GetStock(r.PathValue("product_id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, map[string]int{"quantity": stock})
}

func (s *Server) listBatches(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.BatchFilter{ProductID: r.URL.Query().Get("product_id")}
	items, total, err := s.svc.ListBatches(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{Items: items, Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total}})
}

func (s *Server) listMovements(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.MovementFilter{
		ProductID: r.URL.Query().Get("product_id"),
		Type:      r.URL.Query().Get("type"),
	}
	items, total, err := s.svc.ListMovements(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{Items: items, Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total}})
}
