package handler

import (
	"net/http"

	"supplychain/internal/model"
	"supplychain/pkg/httpx"
)

func (s *Server) registerImportExportRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/products/import", s.importProducts)
	mux.HandleFunc("GET /api/export", s.export)
}

type importProductItem struct {
	SKU               string `json:"sku"`
	Name              string `json:"name"`
	Category          string `json:"category"`
	Unit              string `json:"unit"`
	Price             int64  `json:"price"`
	LowStockThreshold int    `json:"low_stock_threshold"`
}

type importProductsRequest struct {
	Products []importProductItem `json:"products"`
}

func (s *Server) importProducts(w http.ResponseWriter, r *http.Request) {
	var req importProductsRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.BadRequest(w, "请求体解析失败: "+err.Error())
		return
	}
	if len(req.Products) == 0 {
		httpx.BadRequest(w, "商品列表不能为空")
		return
	}
	inputs := make([]model.Product, 0, len(req.Products))
	for _, item := range req.Products {
		inputs = append(inputs, model.Product{
			SKU: item.SKU, Name: item.Name, Category: item.Category, Unit: item.Unit,
			Price: item.Price, LowStockThreshold: item.LowStockThreshold,
		})
	}
	results, err := s.svc.ImportProducts(inputs)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, results)
}

func (s *Server) export(w http.ResponseWriter, r *http.Request) {
	snapshot, err := s.svc.Export()
	if err != nil {
		writeServiceError(w, err)
		return
	}
	httpx.OK(w, snapshot)
}
