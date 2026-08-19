// Package store 定义数据访问接口与内存实现。
package store

import (
	"errors"

	"supplychain/internal/model"
)

var (
	ErrNotFound = errors.New("记录不存在")
	ErrConflict = errors.New("记录已存在或状态冲突")
)

// Store 聚合全部实体的数据访问方法。
type Store interface {
	CreateSupplier(s *model.Supplier) error
	GetSupplier(id string) (*model.Supplier, error)
	ListSuppliers() []*model.Supplier
	UpdateSupplier(s *model.Supplier) error
	DeleteSupplier(id string) error

	CreateProduct(p *model.Product) error
	GetProduct(id string) (*model.Product, error)
	GetProductBySKU(sku string) (*model.Product, error)
	ListProducts() []*model.Product
	UpdateProduct(p *model.Product) error
	DeleteProduct(id string) error

	CreatePurchaseOrder(p *model.PurchaseOrder) error
	GetPurchaseOrder(id string) (*model.PurchaseOrder, error)
	ListPurchaseOrders() []*model.PurchaseOrder
	UpdatePurchaseOrder(p *model.PurchaseOrder) error
	DeletePurchaseOrder(id string) error

	CreateInboundOrder(i *model.InboundOrder) error
	GetInboundOrder(id string) (*model.InboundOrder, error)
	ListInboundOrders() []*model.InboundOrder
	ListInboundOrdersByPurchaseOrder(poID string) []*model.InboundOrder
	UpdateInboundOrder(i *model.InboundOrder) error
	DeleteInboundOrder(id string) error

	CreateInspection(i *model.Inspection) error
	GetInspection(id string) (*model.Inspection, error)
	ListInspections() []*model.Inspection
	GetInspectionByInbound(inboundID string) (*model.Inspection, error)
	UpdateInspection(i *model.Inspection) error
	DeleteInspection(id string) error

	CreateBatch(b *model.InventoryBatch) error
	GetBatch(id string) (*model.InventoryBatch, error)
	ListBatches() []*model.InventoryBatch
	ListBatchesByProduct(productID string) []*model.InventoryBatch
	ListBatchesForOutbound(productID string) []*model.InventoryBatch
	UpdateBatch(b *model.InventoryBatch) error
	DeleteBatch(id string) error

	CreateMovement(m *model.StockMovement) error
	GetMovement(id string) (*model.StockMovement, error)
	ListMovements() []*model.StockMovement
	ListMovementsByProduct(productID string) []*model.StockMovement
	DeleteMovement(id string) error

	CreateReturnOrder(r *model.ReturnOrder) error
	GetReturnOrder(id string) (*model.ReturnOrder, error)
	ListReturnOrders() []*model.ReturnOrder
	ListReturnOrdersByInbound(inboundID string) []*model.ReturnOrder
	UpdateReturnOrder(r *model.ReturnOrder) error
	DeleteReturnOrder(id string) error
}
