package store

import (
	"sync"

	"supplychain/internal/model"
)

// MemoryStore 基于内存的 Store 实现，线程安全。
type MemoryStore struct {
	mu         sync.RWMutex
	suppliers  map[string]*model.Supplier
	products   map[string]*model.Product
	pos        map[string]*model.PurchaseOrder
	inbounds   map[string]*model.InboundOrder
	inspections map[string]*model.Inspection
	batches    map[string]*model.InventoryBatch
	movements  map[string]*model.StockMovement
	returns    map[string]*model.ReturnOrder
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		suppliers:   make(map[string]*model.Supplier),
		products:    make(map[string]*model.Product),
		pos:         make(map[string]*model.PurchaseOrder),
		inbounds:    make(map[string]*model.InboundOrder),
		inspections: make(map[string]*model.Inspection),
		batches:     make(map[string]*model.InventoryBatch),
		movements:   make(map[string]*model.StockMovement),
		returns:     make(map[string]*model.ReturnOrder),
	}
}

var _ Store = (*MemoryStore)(nil)
