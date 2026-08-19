package handler

import (
	"net/http"

	"supplychain/internal/model"
	"supplychain/pkg/httpx"
)

func (s *Server) registerProductRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/products", s.createProduct)
	mux.HandleFunc("GET /api/products", s.listProducts)
	mux.HandleFunc("GET /api/products/{id}", s.getProduct)
	mux.HandleFunc("PUT /api/products/{id}", s.updateProduct)
	mux.HandleFunc("DELETE /api/products/{id}", s.deleteProduct)
}

type productRequest struct {
	SKU               string `json:"sku"`
	Name              string `json:"name"`
	Category          string `json:"category"`
	Unit              string `json:"unit"`
	Price             int64  `json:"price"`
	LowStockThreshold int    `json:"low_stock_threshold"`
}

func (s *Server) createProduct(w http.ResponseWriter, r *http.Request) {
	var req productRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	p, err := s.svc.CreateProduct(model.Product{
		SKU: req.SKU, Name: req.Name, Category: req.Category, Unit: req.Unit,
		Price: req.Price, LowStockThreshold: req.LowStockThreshold,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.Created(w, p)
}

func (s *Server) listProducts(w http.ResponseWriter, r *http.Request) {
	pp := httpx.ParsePagination(r, 20, s.maxPageSize())
	filter := model.ProductFilter{Category: r.URL.Query().Get("category"), Keyword: r.URL.Query().Get("keyword")}
	items, total, err := s.svc.ListProducts(filter, pp.Page, pp.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, httpx.PageResult{Items: items, Pagination: httpx.Pagination{Page: pp.Page, Size: pp.Size, Total: total}})
}

func (s *Server) getProduct(w http.ResponseWriter, r *http.Request) {
	p, err := s.svc.GetProduct(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, p)
}

func (s *Server) updateProduct(w http.ResponseWriter, r *http.Request) {
	var req productRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	p, err := s.svc.UpdateProduct(r.PathValue("id"), model.Product{
		SKU: req.SKU, Name: req.Name, Category: req.Category, Unit: req.Unit,
		Price: req.Price, LowStockThreshold: req.LowStockThreshold,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, p)
}

func (s *Server) deleteProduct(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteProduct(r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.NoContent(w)
}
