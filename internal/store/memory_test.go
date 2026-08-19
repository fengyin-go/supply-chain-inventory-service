package store

import (
	"errors"
	"testing"

	"supplychain/internal/model"
)

func TestSupplierStoreCRUD(t *testing.T) {
	s := NewMemoryStore()

	sup := &model.Supplier{ID: "s1", Name: "华东电子", Phone: "13800000001"}
	if err := s.CreateSupplier(sup); err != nil {
		t.Fatalf("CreateSupplier: %v", err)
	}
	if err := s.CreateSupplier(&model.Supplier{ID: "s2", Name: "华东电子"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}

	got, err := s.GetSupplier("s1")
	if err != nil {
		t.Fatalf("GetSupplier: %v", err)
	}
	if got.Name != "华东电子" {
		t.Fatalf("name = %q", got.Name)
	}

	if _, err := s.GetSupplier("missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if list := s.ListSuppliers(); len(list) != 1 {
		t.Fatalf("len = %d", len(list))
	}

	sup.Name = "华东电子科技"
	if err := s.UpdateSupplier(sup); err != nil {
		t.Fatalf("UpdateSupplier: %v", err)
	}
	if err := s.DeleteSupplier("s1"); err != nil {
		t.Fatalf("DeleteSupplier: %v", err)
	}
	if err := s.DeleteSupplier("s1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestProductStoreCRUD(t *testing.T) {
	s := NewMemoryStore()

	p := &model.Product{ID: "p1", SKU: "SKU001", Name: "螺丝"}
	if err := s.CreateProduct(p); err != nil {
		t.Fatalf("CreateProduct: %v", err)
	}
	if err := s.CreateProduct(&model.Product{ID: "p2", SKU: "SKU001"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}

	if got, err := s.GetProductBySKU("SKU001"); err != nil || got.ID != "p1" {
		t.Fatalf("GetProductBySKU: %v %v", got, err)
	}
	if _, err := s.GetProductBySKU("NOPE"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if list := s.ListProducts(); len(list) != 1 {
		t.Fatalf("len = %d", len(list))
	}

	p.Name = "螺母"
	if err := s.UpdateProduct(p); err != nil {
		t.Fatalf("UpdateProduct: %v", err)
	}
	if err := s.DeleteProduct("p1"); err != nil {
		t.Fatalf("DeleteProduct: %v", err)
	}
}

func TestPurchaseOrderStoreCRUD(t *testing.T) {
	s := NewMemoryStore()

	po := &model.PurchaseOrder{ID: "po1", OrderNo: "PO-1", SupplierID: "s1", Status: model.POStatusDraft}
	if err := s.CreatePurchaseOrder(po); err != nil {
		t.Fatalf("CreatePurchaseOrder: %v", err)
	}
	if err := s.CreatePurchaseOrder(&model.PurchaseOrder{ID: "po2", OrderNo: "PO-1"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}

	got, err := s.GetPurchaseOrder("po1")
	if err != nil {
		t.Fatalf("GetPurchaseOrder: %v", err)
	}
	if got.Status != model.POStatusDraft {
		t.Fatalf("status = %q", got.Status)
	}

	if list := s.ListPurchaseOrders(); len(list) != 1 {
		t.Fatalf("len = %d", len(list))
	}

	po.Status = model.POStatusConfirmed
	if err := s.UpdatePurchaseOrder(po); err != nil {
		t.Fatalf("UpdatePurchaseOrder: %v", err)
	}
	if err := s.DeletePurchaseOrder("po1"); err != nil {
		t.Fatalf("DeletePurchaseOrder: %v", err)
	}
}

func TestInboundOrderStoreCRUD(t *testing.T) {
	s := NewMemoryStore()

	in := &model.InboundOrder{ID: "in1", InboundNo: "IN-1", PurchaseOrderID: "po1", Status: model.InboundPending}
	if err := s.CreateInboundOrder(in); err != nil {
		t.Fatalf("CreateInboundOrder: %v", err)
	}

	got, err := s.GetInboundOrder("in1")
	if err != nil {
		t.Fatalf("GetInboundOrder: %v", err)
	}
	if got.PurchaseOrderID != "po1" {
		t.Fatalf("po = %q", got.PurchaseOrderID)
	}

	byPO := s.ListInboundOrdersByPurchaseOrder("po1")
	if len(byPO) != 1 {
		t.Fatalf("len = %d", len(byPO))
	}
	if len(s.ListInboundOrdersByPurchaseOrder("poX")) != 0 {
		t.Fatal("expected empty")
	}

	in.Status = model.InboundInspecting
	if err := s.UpdateInboundOrder(in); err != nil {
		t.Fatalf("UpdateInboundOrder: %v", err)
	}
	if err := s.DeleteInboundOrder("in1"); err != nil {
		t.Fatalf("DeleteInboundOrder: %v", err)
	}
}

func TestInspectionStoreCRUD(t *testing.T) {
	s := NewMemoryStore()

	ins := &model.Inspection{ID: "i1", InboundOrderID: "in1", Result: model.InspectionPending}
	if err := s.CreateInspection(ins); err != nil {
		t.Fatalf("CreateInspection: %v", err)
	}

	got, err := s.GetInspectionByInbound("in1")
	if err != nil {
		t.Fatalf("GetInspectionByInbound: %v", err)
	}
	if got.ID != "i1" {
		t.Fatalf("id = %q", got.ID)
	}
	if _, err := s.GetInspectionByInbound("inX"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	if list := s.ListInspections(); len(list) != 1 {
		t.Fatalf("len = %d", len(list))
	}

	ins.Result = model.InspectionPassed
	if err := s.UpdateInspection(ins); err != nil {
		t.Fatalf("UpdateInspection: %v", err)
	}
	if err := s.DeleteInspection("i1"); err != nil {
		t.Fatalf("DeleteInspection: %v", err)
	}
}

func TestBatchStoreCRUD(t *testing.T) {
	s := NewMemoryStore()

	b := &model.InventoryBatch{ID: "b1", BatchNo: "BT-1", ProductID: "p1", Quantity: 10, Remaining: 10}
	if err := s.CreateBatch(b); err != nil {
		t.Fatalf("CreateBatch: %v", err)
	}

	got, err := s.GetBatch("b1")
	if err != nil {
		t.Fatalf("GetBatch: %v", err)
	}
	if got.Remaining != 10 {
		t.Fatalf("remaining = %d", got.Remaining)
	}

	byProduct := s.ListBatchesByProduct("p1")
	if len(byProduct) != 1 {
		t.Fatalf("len = %d", len(byProduct))
	}

	b.Remaining = 5
	if err := s.UpdateBatch(b); err != nil {
		t.Fatalf("UpdateBatch: %v", err)
	}
	if err := s.DeleteBatch("b1"); err != nil {
		t.Fatalf("DeleteBatch: %v", err)
	}
}

func TestMovementStoreCRUD(t *testing.T) {
	s := NewMemoryStore()

	m := &model.StockMovement{ID: "m1", ProductID: "p1", Type: model.MovementInbound, Delta: 10}
	if err := s.CreateMovement(m); err != nil {
		t.Fatalf("CreateMovement: %v", err)
	}

	got, err := s.GetMovement("m1")
	if err != nil {
		t.Fatalf("GetMovement: %v", err)
	}
	if got.Delta != 10 {
		t.Fatalf("delta = %d", got.Delta)
	}

	byProduct := s.ListMovementsByProduct("p1")
	if len(byProduct) != 1 {
		t.Fatalf("len = %d", len(byProduct))
	}

	if list := s.ListMovements(); len(list) != 1 {
		t.Fatalf("len = %d", len(list))
	}
	if err := s.DeleteMovement("m1"); err != nil {
		t.Fatalf("DeleteMovement: %v", err)
	}
}

func TestReturnOrderStoreCRUD(t *testing.T) {
	s := NewMemoryStore()

	r := &model.ReturnOrder{ID: "r1", ReturnNo: "RT-1", InboundOrderID: "in1", ProductID: "p1", Quantity: 2, Status: model.ReturnPending}
	if err := s.CreateReturnOrder(r); err != nil {
		t.Fatalf("CreateReturnOrder: %v", err)
	}
	if err := s.CreateReturnOrder(&model.ReturnOrder{ID: "r2", ReturnNo: "RT-1"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}

	got, err := s.GetReturnOrder("r1")
	if err != nil {
		t.Fatalf("GetReturnOrder: %v", err)
	}
	if got.Quantity != 2 {
		t.Fatalf("quantity = %d", got.Quantity)
	}

	byInbound := s.ListReturnOrdersByInbound("in1")
	if len(byInbound) != 1 {
		t.Fatalf("len = %d", len(byInbound))
	}

	r.Status = model.ReturnCompleted
	if err := s.UpdateReturnOrder(r); err != nil {
		t.Fatalf("UpdateReturnOrder: %v", err)
	}
	if err := s.DeleteReturnOrder("r1"); err != nil {
		t.Fatalf("DeleteReturnOrder: %v", err)
	}
}
