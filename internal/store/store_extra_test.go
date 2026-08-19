package store

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"supplychain/internal/model"
)

func TestUpdateNotFound(t *testing.T) {
	s := NewMemoryStore()

	if err := s.UpdateSupplier(&model.Supplier{ID: "missing"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("UpdateSupplier: %v", err)
	}
	if err := s.UpdateProduct(&model.Product{ID: "missing"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("UpdateProduct: %v", err)
	}
	if err := s.UpdatePurchaseOrder(&model.PurchaseOrder{ID: "missing"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("UpdatePurchaseOrder: %v", err)
	}
	if err := s.UpdateInboundOrder(&model.InboundOrder{ID: "missing"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("UpdateInboundOrder: %v", err)
	}
	if err := s.UpdateInspection(&model.Inspection{ID: "missing"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("UpdateInspection: %v", err)
	}
	if err := s.UpdateBatch(&model.InventoryBatch{ID: "missing"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("UpdateBatch: %v", err)
	}
	if err := s.UpdateReturnOrder(&model.ReturnOrder{ID: "missing"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("UpdateReturnOrder: %v", err)
	}
}

func TestUpdateConflictScenarios(t *testing.T) {
	s := NewMemoryStore()
	s.CreateSupplier(&model.Supplier{ID: "s1", Name: "A"})
	s.CreateSupplier(&model.Supplier{ID: "s2", Name: "B"})

	// 将 s2 改名为 A → 冲突
	if err := s.UpdateSupplier(&model.Supplier{ID: "s2", Name: "A"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}

	s.CreateProduct(&model.Product{ID: "p1", SKU: "SKU1"})
	s.CreateProduct(&model.Product{ID: "p2", SKU: "SKU2"})
	if err := s.UpdateProduct(&model.Product{ID: "p2", SKU: "SKU1"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := NewMemoryStore()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = s.CreateSupplier(&model.Supplier{ID: fmt.Sprintf("s%d", n), Name: fmt.Sprintf("供应商%d", n), Phone: "13800000001"})
			_ = s.ListSuppliers()
			_, _ = s.GetSupplier(fmt.Sprintf("s%d", n))
		}(i)
	}
	wg.Wait()

	list := s.ListSuppliers()
	if len(list) != 100 {
		t.Fatalf("expected 100 suppliers, got %d", len(list))
	}
}

func TestConcurrentMovementCreation(t *testing.T) {
	s := NewMemoryStore()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = s.CreateMovement(&model.StockMovement{ID: fmt.Sprintf("m%d", n), ProductID: "p1", Type: model.MovementInbound, Delta: 1})
		}(i)
	}
	wg.Wait()

	if list := s.ListMovements(); len(list) != 50 {
		t.Fatalf("expected 50 movements, got %d", len(list))
	}
}
