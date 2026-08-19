package handler

import (
	"net/http"

	"supplychain/internal/model"
	"supplychain/pkg/httpx"
)

func (s *Server) registerInboundRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/inbound-orders", s.createInboundOrder)
	mux.HandleFunc("GET /api/inbound-orders", s.listInboundOrders)
	mux.HandleFunc("GET /api/inbound-orders/{id}", s.getInboundOrder)
	mux.HandleFunc("POST /api/inbound-orders/{id}/reject", s.rejectInbound)
	mux.HandleFunc("DELETE /api/inbound-orders/{id}", s.deleteInboundOrder)
}

type createInboundRequest struct {
	PurchaseOrderID string `json:"purchase_order_id"`
}

func (s *Server) createInboundOrder(w http.ResponseWriter, r *http.Request) {
	var req createInboundRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	inbound, err := s.svc.CreateInboundOrder(req.PurchaseOrderID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, inbound)
}

func (s *Server) listInboundOrders(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.InboundFilter{
		Status:          r.URL.Query().Get("status"),
		PurchaseOrderID: r.URL.Query().Get("purchase_order_id"),
	}
	items, total, err := s.svc.ListInboundOrders(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{Items: items, Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total}})
}

func (s *Server) getInboundOrder(w http.ResponseWriter, r *http.Request) {
	inbound, err := s.svc.GetInboundOrder(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, inbound)
}

func (s *Server) rejectInbound(w http.ResponseWriter, r *http.Request) {
	inbound, err := s.svc.RejectInbound(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, inbound)
}

func (s *Server) deleteInboundOrder(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteInboundOrder(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}
