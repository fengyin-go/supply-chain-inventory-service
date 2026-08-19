package service

import (
	"testing"

	"supplychain/internal/model"
)

func seedPurchaseOrder(t *testing.T, s *Service, supplierID string, items []model.PurchaseItem) *model.PurchaseOrder {
	t.Helper()
	po, err := s.CreatePurchaseOrder(model.PurchaseOrder{SupplierID: supplierID, Items: items})
	if err != nil {
		t.Fatalf("CreatePurchaseOrder: %v", err)
	}
	return po
}

func TestPurchaseOrderLifecycle(t *testing.T) {
	s := newTestService()
	sup := seedSupplier(t, s, "华东电子", "13800000001")
	p1 := seedProduct(t, s, "SKU001", "螺丝", 100, 10)
	p2 := seedProduct(t, s, "SKU002", "螺母", 200, 5)

	po := seedPurchaseOrder(t, s, sup.ID, []model.PurchaseItem{
		{ProductID: p1.ID, Quantity: 10, UnitPrice: 100},
		{ProductID: p2.ID, Quantity: 5, UnitPrice: 200},
	})
	if po.Status != model.POStatusDraft {
		t.Fatalf("status = %q, want draft", po.Status)
	}
	if po.OrderNo == "" {
		t.Fatal("expected generated order no")
	}
	if po.TotalAmount != 2000 {
		t.Fatalf("total = %d, want 2000", po.TotalAmount)
	}

	// 确认
	confirmed, err := s.ConfirmPurchaseOrder(po.ID)
	if err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	if confirmed.Status != model.POStatusConfirmed {
		t.Fatalf("status = %q", confirmed.Status)
	}

	// 再次确认失败
	if _, err := s.ConfirmPurchaseOrder(po.ID); !model.IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestPurchaseOrderCrossEntityValidation(t *testing.T) {
	s := newTestService()
	p := seedProduct(t, s, "SKU001", "螺丝", 100, 10)

	// 不存在的供应商
	if _, err := s.CreatePurchaseOrder(model.PurchaseOrder{
		SupplierID: "missing",
		Items:      []model.PurchaseItem{{ProductID: p.ID, Quantity: 1, UnitPrice: 100}},
	}); err == nil {
		t.Fatal("expected error for missing supplier")
	}

	// 不存在的商品
	sup := seedSupplier(t, s, "华东电子", "13800000001")
	if _, err := s.CreatePurchaseOrder(model.PurchaseOrder{
		SupplierID: sup.ID,
		Items:      []model.PurchaseItem{{ProductID: "missing", Quantity: 1, UnitPrice: 100}},
	}); err == nil {
		t.Fatal("expected error for missing product")
	}
}

func TestPurchaseOrderCancel(t *testing.T) {
	s := newTestService()
	sup := seedSupplier(t, s, "华东电子", "13800000001")
	p := seedProduct(t, s, "SKU001", "螺丝", 100, 10)
	po := seedPurchaseOrder(t, s, sup.ID, []model.PurchaseItem{{ProductID: p.ID, Quantity: 1, UnitPrice: 100}})

	cancelled, err := s.CancelPurchaseOrder(po.ID)
	if err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if cancelled.Status != model.POStatusCancelled {
		t.Fatalf("status = %q", cancelled.Status)
	}

	// 取消后不可确认
	if _, err := s.ConfirmPurchaseOrder(po.ID); !model.IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestListPurchaseOrdersFilter(t *testing.T) {
	s := newTestService()
	sup := seedSupplier(t, s, "华东电子", "13800000001")
	p := seedProduct(t, s, "SKU001", "螺丝", 100, 10)
	po1 := seedPurchaseOrder(t, s, sup.ID, []model.PurchaseItem{{ProductID: p.ID, Quantity: 1, UnitPrice: 100}})
	po2 := seedPurchaseOrder(t, s, sup.ID, []model.PurchaseItem{{ProductID: p.ID, Quantity: 2, UnitPrice: 100}})
	s.ConfirmPurchaseOrder(po1.ID)

	_, total, _ := s.ListPurchaseOrders(model.PurchaseOrderFilter{Status: model.POStatusConfirmed}, 1, 10)
	if total != 1 {
		t.Fatalf("confirmed total = %d, want 1", total)
	}
	_, total, _ = s.ListPurchaseOrders(model.PurchaseOrderFilter{SupplierID: sup.ID}, 1, 10)
	if total != 2 {
		t.Fatalf("supplier total = %d, want 2", total)
	}
	// 分页
	items, total, _ := s.ListPurchaseOrders(model.PurchaseOrderFilter{}, 1, 1)
	if total != 2 || len(items) != 1 {
		t.Fatalf("page total=%d len=%d", total, len(items))
	}
	_ = po2
}
