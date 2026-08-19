package handler

import (
	"net/http"

	"supplychain/internal/model"
	"supplychain/pkg/httpx"
)

func (s *Server) registerReturnRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/return-orders", s.createReturnOrder)
	mux.HandleFunc("GET /api/return-orders", s.listReturnOrders)
	mux.HandleFunc("GET /api/return-orders/{id}", s.getReturnOrder)
	mux.HandleFunc("POST /api/return-orders/{id}/complete", s.completeReturnOrder)
	mux.HandleFunc("GET /api/reports/returns", s.returnStats)
	mux.HandleFunc("GET /api/batches/{id}/trace", s.traceBatch)
}

type createReturnRequest struct {
	InboundOrderID string `json:"inbound_order_id"`
	ProductID      string `json:"product_id"`
	Quantity       int    `json:"quantity"`
	Reason         string `json:"reason"`
}

func (s *Server) createReturnOrder(w http.ResponseWriter, r *http.Request) {
	var req createReturnRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	ret, err := s.svc.CreateReturnOrder(req.InboundOrderID, req.ProductID, req.Quantity, req.Reason)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, ret)
}

func (s *Server) listReturnOrders(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.ReturnFilter{
		Status:         r.URL.Query().Get("status"),
		InboundOrderID: r.URL.Query().Get("inbound_order_id"),
	}
	items, total, err := s.svc.ListReturnOrders(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{Items: items, Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total}})
}

func (s *Server) getReturnOrder(w http.ResponseWriter, r *http.Request) {
	ret, err := s.svc.GetReturnOrder(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, ret)
}

func (s *Server) completeReturnOrder(w http.ResponseWriter, r *http.Request) {
	ret, err := s.svc.CompleteReturnOrder(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, ret)
}

func (s *Server) returnStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.svc.ReturnStats()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, stats)
}

func (s *Server) traceBatch(w http.ResponseWriter, r *http.Request) {
	trace, err := s.svc.TraceBatch(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, trace)
}
