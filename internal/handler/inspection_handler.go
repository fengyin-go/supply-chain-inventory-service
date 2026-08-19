package handler

import (
	"net/http"

	"supplychain/internal/model"
	"supplychain/pkg/httpx"
)

func (s *Server) registerInspectionRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/inspections", s.startInspection)
	mux.HandleFunc("GET /api/inspections", s.listInspections)
	mux.HandleFunc("GET /api/inspections/{id}", s.getInspection)
	mux.HandleFunc("POST /api/inspections/{id}/complete", s.completeInspection)
}

type startInspectionRequest struct {
	InboundOrderID string `json:"inbound_order_id"`
	Inspector      string `json:"inspector"`
}

func (s *Server) startInspection(w http.ResponseWriter, r *http.Request) {
	var req startInspectionRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	ins, err := s.svc.StartInspection(req.InboundOrderID, req.Inspector)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, ins)
}

func (s *Server) listInspections(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.InspectionFilter{
		Result:         r.URL.Query().Get("result"),
		InboundOrderID: r.URL.Query().Get("inbound_order_id"),
	}
	items, total, err := s.svc.ListInspections(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{Items: items, Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total}})
}

func (s *Server) getInspection(w http.ResponseWriter, r *http.Request) {
	ins, err := s.svc.GetInspection(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, ins)
}

type completeInspectionRequest struct {
	Result      string `json:"result"`
	DefectCount int    `json:"defect_count"`
	Note        string `json:"note"`
}

func (s *Server) completeInspection(w http.ResponseWriter, r *http.Request) {
	var req completeInspectionRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	ins, err := s.svc.CompleteInspection(r.PathValue("id"), req.Result, req.DefectCount, req.Note)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, ins)
}
