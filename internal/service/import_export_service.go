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

// Export 导出当前内存中的全量数据快照。
// 返回的是深拷贝副本，调用方对快照的任何修改都不会回写到内存中的原始记录。
func (s *Service) Export() (*ExportSnapshot, error) {
	return &ExportSnapshot{
		Suppliers:      model.CloneSlice(s.store.ListSuppliers()),
		Products:       model.CloneSlice(s.store.ListProducts()),
		PurchaseOrders: model.CloneSlice(s.store.ListPurchaseOrders()),
		InboundOrders:  model.CloneSlice(s.store.ListInboundOrders()),
		Inspections:    model.CloneSlice(s.store.ListInspections()),
		Batches:        model.CloneSlice(s.store.ListBatches()),
		Movements:      model.CloneSlice(s.store.ListMovements()),
		ExportedAt:     time.Now(),
	}, nil
}
