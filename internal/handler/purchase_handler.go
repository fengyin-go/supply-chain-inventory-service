package handler

import (
	"net/http"

	"supplychain/internal/model"
	"supplychain/pkg/httpx"
)

func (s *Server) registerPurchaseRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/purchase-orders", s.createPurchaseOrder)
	mux.HandleFunc("GET /api/purchase-orders", s.listPurchaseOrders)
	mux.HandleFunc("GET /api/purchase-orders/{id}", s.getPurchaseOrder)
	mux.HandleFunc("POST /api/purchase-orders/{id}/confirm", s.confirmPurchaseOrder)
	mux.HandleFunc("POST /api/purchase-orders/{id}/cancel", s.cancelPurchaseOrder)
	mux.HandleFunc("DELETE /api/purchase-orders/{id}", s.deletePurchaseOrder)
}

type purchaseItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	UnitPrice int64  `json:"unit_price"`
}

type purchaseOrderRequest struct {
	OrderNo    string                `json:"order_no"`
	SupplierID string                `json:"supplier_id"`
	Items      []purchaseItemRequest `json:"items"`
}

func (s *Server) createPurchaseOrder(w http.ResponseWriter, r *http.Request) {
	var req purchaseOrderRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	items := make([]model.PurchaseItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, model.PurchaseItem{ProductID: it.ProductID, Quantity: it.Quantity, UnitPrice: it.UnitPrice})
	}
	po, err := s.svc.CreatePurchaseOrder(model.PurchaseOrder{
		OrderNo: req.OrderNo, SupplierID: req.SupplierID, Items: items,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, po)
}

func (s *Server) listPurchaseOrders(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.PurchaseOrderFilter{
		Status:     r.URL.Query().Get("status"),
		SupplierID: r.URL.Query().Get("supplier_id"),
	}
	items, total, err := s.svc.ListPurchaseOrders(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{Items: items, Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total}})
}

func (s *Server) getPurchaseOrder(w http.ResponseWriter, r *http.Request) {
	po, err := s.svc.GetPurchaseOrder(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, po)
}

func (s *Server) confirmPurchaseOrder(w http.ResponseWriter, r *http.Request) {
	po, err := s.svc.ConfirmPurchaseOrder(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, po)
}

func (s *Server) cancelPurchaseOrder(w http.ResponseWriter, r *http.Request) {
	po, err := s.svc.CancelPurchaseOrder(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, po)
}

func (s *Server) deletePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeletePurchaseOrder(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}
