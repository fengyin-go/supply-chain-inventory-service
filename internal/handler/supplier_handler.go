package handler

import (
	"net/http"

	"supplychain/internal/model"
	"supplychain/pkg/httpx"
)

func (s *Server) registerSupplierRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/suppliers", s.createSupplier)
	mux.HandleFunc("GET /api/suppliers", s.listSuppliers)
	mux.HandleFunc("GET /api/suppliers/{id}", s.getSupplier)
	mux.HandleFunc("PUT /api/suppliers/{id}", s.updateSupplier)
	mux.HandleFunc("DELETE /api/suppliers/{id}", s.deleteSupplier)
}

type supplierRequest struct {
	Name    string `json:"name"`
	Contact string `json:"contact"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
	Status  string `json:"status"`
}

func (s *Server) createSupplier(w http.ResponseWriter, r *http.Request) {
	var req supplierRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	sup, err := s.svc.CreateSupplier(model.Supplier{
		Name: req.Name, Contact: req.Contact, Phone: req.Phone, Address: req.Address, Status: req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, sup)
}

func (s *Server) listSuppliers(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.SupplierFilter{Status: r.URL.Query().Get("status"), Keyword: r.URL.Query().Get("keyword")}
	items, total, err := s.svc.ListSuppliers(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{Items: items, Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total}})
}

func (s *Server) getSupplier(w http.ResponseWriter, r *http.Request) {
	sup, err := s.svc.GetSupplier(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, sup)
}

func (s *Server) updateSupplier(w http.ResponseWriter, r *http.Request) {
	var req supplierRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	sup, err := s.svc.UpdateSupplier(r.PathValue("id"), model.Supplier{
		Name: req.Name, Contact: req.Contact, Phone: req.Phone, Address: req.Address, Status: req.Status,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, sup)
}

func (s *Server) deleteSupplier(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteSupplier(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}
