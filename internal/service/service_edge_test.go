package service

import (
	"testing"

	"supplychain/internal/model"
)

func TestListFiltersAndPagination(t *testing.T) {
	s := newTestService()
	sup := seedSupplier(t, s, "华东电子", "13800000001")
	p1 := seedProduct(t, s, "SKU001", "螺丝", 100, 10)
	p2 := seedProduct(t, s, "SKU002", "螺母", 200, 5)
	po := seedPurchaseOrder(t, s, sup.ID, []model.PurchaseItem{
		{ProductID: p1.ID, Quantity: 10, UnitPrice: 100},
		{ProductID: p2.ID, Quantity: 5, UnitPrice: 200},
	})
	s.ConfirmPurchaseOrder(po.ID)

	// 商品分页
	_, total, _ := s.ListProducts(model.ProductFilter{}, 1, 1)
	if total != 2 {
		t.Fatalf("product total = %d, want 2", total)
	}
	items, _, _ := s.ListProducts(model.ProductFilter{}, 1, 1)
	if len(items) != 1 {
		t.Fatalf("product page len = %d, want 1", len(items))
	}
	// 越界页
	items, _, _ = s.ListProducts(model.ProductFilter{}, 99, 1)
	if len(items) != 0 {
		t.Fatalf("empty page len = %d", len(items))
	}

	// 供应商分页
	suppliers, _, _ := s.ListSuppliers(model.SupplierFilter{}, 1, 10)
	if len(suppliers) != 1 {
		t.Fatalf("supplier len = %d", len(suppliers))
	}
	// 无匹配筛选
	_, total, _ = s.ListSuppliers(model.SupplierFilter{Keyword: "不存在"}, 1, 10)
	if total != 0 {
		t.Fatalf("no-match total = %d", total)
	}
}

func TestStockEdgeCases(t *testing.T) {
	s := newTestService()
	p := seedProduct(t, s, "SKU001", "螺丝", 100, 10)

	// 未入库即出库 → 库存不足
	if err := s.OutboundStock(p.ID, 1, "出库"); !model.IsValidationError(err) {
		t.Fatalf("expected ValidationError for empty stock, got %v", err)
	}

	// 调整 delta 为 0
	if err := s.AdjustStock(p.ID, 0, "零调整"); !model.IsValidationError(err) {
		t.Fatalf("expected ValidationError for zero adjust, got %v", err)
	}

	// 负向调整不存在的商品
	if err := s.AdjustStock("missing", -1, "x"); err == nil {
		t.Fatal("expected error for missing product adjust")
	}
}

func TestConflictScenarios(t *testing.T) {
	s := newTestService()
	seedSupplier(t, s, "华东电子", "13800000001")
	seedProduct(t, s, "SKU001", "螺丝", 100, 10)

	// 供应商重名
	if _, err := s.CreateSupplier(model.Supplier{Name: "华东电子", Phone: "13800000002"}); err == nil {
		t.Fatal("expected conflict on duplicate supplier name")
	}
	// 商品重 SKU
	if _, err := s.CreateProduct(model.Product{SKU: "SKU001", Name: "另一螺丝"}); err == nil {
		t.Fatal("expected conflict on duplicate SKU")
	}
}

func TestDeleteNonexistent(t *testing.T) {
	s := newTestService()
	if err := s.DeleteSupplier("missing"); err == nil {
		t.Fatal("expected error deleting missing supplier")
	}
	if err := s.DeleteProduct("missing"); err == nil {
		t.Fatal("expected error deleting missing product")
	}
	if err := s.DeletePurchaseOrder("missing"); err == nil {
		t.Fatal("expected error deleting missing po")
	}
	if err := s.DeleteInboundOrder("missing"); err == nil {
		t.Fatal("expected error deleting missing inbound")
	}
}

func TestUpdateSupplierConflict(t *testing.T) {
	s := newTestService()
	seedSupplier(t, s, "供应商A", "13800000001")
	s2 := seedSupplier(t, s, "供应商B", "13800000002")

	// 将 B 改名为 A → 冲突
	if _, err := s.UpdateSupplier(s2.ID, model.Supplier{Name: "供应商A", Contact: "x", Phone: "13800000002", Address: "y", Status: model.SupplierActive}); err == nil {
		t.Fatal("expected conflict on duplicate name during update")
	}
}

func TestUpdateProductConflict(t *testing.T) {
	s := newTestService()
	seedProduct(t, s, "SKU001", "螺丝", 100, 10)
	p2 := seedProduct(t, s, "SKU002", "螺母", 200, 5)

	// 将 p2 的 SKU 改为 SKU001 → 冲突
	if _, err := s.UpdateProduct(p2.ID, model.Product{SKU: "SKU001", Name: "螺母", Price: 200, LowStockThreshold: 5}); err == nil {
		t.Fatal("expected conflict on duplicate SKU during update")
	}
}
