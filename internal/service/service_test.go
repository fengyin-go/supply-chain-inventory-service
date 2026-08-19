package service

import (
	"testing"

	"supplychain/internal/config"
	"supplychain/internal/model"
	"supplychain/internal/store"
	"supplychain/pkg/logger"
)

func newTestService() *Service {
	cfg := &config.Config{MaxPageSize: 100}
	log := logger.NewLevel(logger.LevelError)
	return New(store.NewMemoryStore(), log, cfg)
}

func seedSupplier(t *testing.T, s *Service, name, phone string) *model.Supplier {
	t.Helper()
	sup, err := s.CreateSupplier(model.Supplier{Name: name, Contact: "联系人", Phone: phone, Address: "地址", Status: model.SupplierActive})
	if err != nil {
		t.Fatalf("CreateSupplier: %v", err)
	}
	return sup
}

func seedProduct(t *testing.T, s *Service, sku, name string, price int64, threshold int) *model.Product {
	t.Helper()
	p, err := s.CreateProduct(model.Product{SKU: sku, Name: name, Category: "原料", Unit: "件", Price: price, LowStockThreshold: threshold})
	if err != nil {
		t.Fatalf("CreateProduct: %v", err)
	}
	return p
}

func TestSupplierService(t *testing.T) {
	s := newTestService()

	sup := seedSupplier(t, s, "华东电子", "13800000001")

	got, err := s.GetSupplier(sup.ID)
	if err != nil {
		t.Fatalf("GetSupplier: %v", err)
	}
	if got.Status != model.SupplierActive {
		t.Fatalf("default status = %q", got.Status)
	}

	// 校验
	if _, err := s.CreateSupplier(model.Supplier{Name: "", Phone: "13800000002"}); !model.IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	// 重名冲突
	if _, err := s.CreateSupplier(model.Supplier{Name: "华东电子", Phone: "13800000002"}); err == nil {
		t.Fatal("expected conflict")
	}

	// 筛选
	seedSupplier(t, s, "华西机械", "13800000003")
	_, total, _ := s.ListSuppliers(model.SupplierFilter{Keyword: "华东"}, 1, 10)
	if total != 1 {
		t.Fatalf("filter total = %d, want 1", total)
	}

	updated, err := s.UpdateSupplier(sup.ID, model.Supplier{Name: "华东电子科技", Contact: "新联系人", Phone: "13800000001", Address: "新地址", Status: model.SupplierInactive})
	if err != nil {
		t.Fatalf("UpdateSupplier: %v", err)
	}
	if updated.Status != model.SupplierInactive {
		t.Fatalf("status = %q", updated.Status)
	}

	if err := s.DeleteSupplier(sup.ID); err != nil {
		t.Fatalf("DeleteSupplier: %v", err)
	}
}

func TestProductService(t *testing.T) {
	s := newTestService()

	p := seedProduct(t, s, "SKU001", "螺丝", 100, 10)

	got, err := s.GetProduct(p.ID)
	if err != nil {
		t.Fatalf("GetProduct: %v", err)
	}
	if got.Unit != "件" {
		t.Fatalf("default unit = %q", got.Unit)
	}

	// SKU 冲突
	if _, err := s.CreateProduct(model.Product{SKU: "SKU001", Name: "另一螺丝"}); err == nil {
		t.Fatal("expected conflict on duplicate SKU")
	}

	// 筛选
	seedProduct(t, s, "SKU002", "螺母", 200, 5)
	_, total, _ := s.ListProducts(model.ProductFilter{Keyword: "螺丝"}, 1, 10)
	if total != 1 {
		t.Fatalf("filter total = %d, want 1", total)
	}

	updated, err := s.UpdateProduct(p.ID, model.Product{SKU: "SKU001", Name: "不锈钢螺丝", Category: "紧固件", Unit: "盒", Price: 150, LowStockThreshold: 20})
	if err != nil {
		t.Fatalf("UpdateProduct: %v", err)
	}
	if updated.Price != 150 {
		t.Fatalf("price = %d", updated.Price)
	}

	if err := s.DeleteProduct(p.ID); err != nil {
		t.Fatalf("DeleteProduct: %v", err)
	}
}
