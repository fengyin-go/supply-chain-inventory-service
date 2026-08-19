package service

import (
	"time"

	"supplychain/internal/model"
)

// ImportProductResult 批量导入商品的单条结果。
type ImportProductResult struct {
	SKU     string `json:"sku"`
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ImportProducts 批量导入商品，返回逐条结果（部分成功不阻断）。
func (s *Service) ImportProducts(products []model.Product) ([]*ImportProductResult, error) {
	results := make([]*ImportProductResult, 0, len(products))
	for _, p := range products {
		res := &ImportProductResult{SKU: p.SKU}
		created, err := s.CreateProduct(p)
		if err != nil {
			res.Error = err.Error()
		} else {
			res.Success = true
			res.ID = created.ID
		}
		results = append(results, res)
	}
	return results, nil
}

// ExportSnapshot 全量数据导出快照。
type ExportSnapshot struct {
	Suppliers   []*model.Supplier       `json:"suppliers"`
	Products    []*model.Product        `json:"products"`
	PurchaseOrders []*model.PurchaseOrder `json:"purchase_orders"`
	InboundOrders []*model.InboundOrder   `json:"inbound_orders"`
	Inspections []*model.Inspection      `json:"inspections"`
	Batches     []*model.InventoryBatch  `json:"batches"`
	Movements   []*model.StockMovement   `json:"movements"`
	ExportedAt  time.Time               `json:"exported_at"`
}

// Export 导出当前内存中的全量数据。
func (s *Service) Export() (*ExportSnapshot, error) {
	suppliers := s.store.ListSuppliers()
	supplierCopies := make([]*model.Supplier, 0, len(suppliers))
	for _, item := range suppliers {
		supplierCopies = append(supplierCopies, item.Snapshot())
	}
	products := s.store.ListProducts()
	productCopies := make([]*model.Product, 0, len(products))
	for _, item := range products {
		productCopies = append(productCopies, item.Snapshot())
	}
	orders := s.store.ListPurchaseOrders()
	orderCopies := make([]*model.PurchaseOrder, 0, len(orders))
	for _, item := range orders {
		orderCopies = append(orderCopies, item.Snapshot())
	}
	inbounds := s.store.ListInboundOrders()
	inboundCopies := make([]*model.InboundOrder, 0, len(inbounds))
	for _, item := range inbounds {
		inboundCopies = append(inboundCopies, item.Snapshot())
	}
	return &ExportSnapshot{
		Suppliers:      supplierCopies,
		Products:       productCopies,
		PurchaseOrders: orderCopies,
		InboundOrders:  inboundCopies,
		Inspections:    s.store.ListInspections(),
		Batches:        s.store.ListBatches(),
		Movements:      s.store.ListMovements(),
		ExportedAt:     time.Now(),
	}, nil
}
