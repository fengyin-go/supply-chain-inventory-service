package service

import (
	"testing"

	"supplychain/internal/model"
)

// setupConfirmedPO 创建供应商、商品、采购单并确认，返回相关对象。
func setupConfirmedPO(t *testing.T, s *Service) (*model.Supplier, *model.Product, *model.PurchaseOrder) {
	t.Helper()
	sup := seedSupplier(t, s, "华东电子", "13800000001")
	p := seedProduct(t, s, "SKU001", "螺丝", 100, 10)
	po := seedPurchaseOrder(t, s, sup.ID, []model.PurchaseItem{{ProductID: p.ID, Quantity: 10, UnitPrice: 100}})
	if _, err := s.ConfirmPurchaseOrder(po.ID); err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	return sup, p, po
}

func TestInboundOrderCreation(t *testing.T) {
	s := newTestService()
	_, _, po := setupConfirmedPO(t, s)

	inbound, err := s.CreateInboundOrder(po.ID)
	if err != nil {
		t.Fatalf("CreateInboundOrder: %v", err)
	}
	if inbound.Status != model.InboundPending {
		t.Fatalf("status = %q, want pending", inbound.Status)
	}
	if inbound.InboundNo == "" {
		t.Fatal("expected generated inbound no")
	}
	if len(inbound.Items) != 1 || inbound.Items[0].Quantity != 10 {
		t.Fatalf("items = %+v", inbound.Items)
	}

	// 重复创建入库单失败
	if _, err := s.CreateInboundOrder(po.ID); !model.IsValidationError(err) {
		t.Fatalf("expected ValidationError for duplicate inbound, got %v", err)
	}
}

func TestInboundOrderRequiresConfirmedPO(t *testing.T) {
	s := newTestService()
	sup := seedSupplier(t, s, "华东电子", "13800000001")
	p := seedProduct(t, s, "SKU001", "螺丝", 100, 10)
	po := seedPurchaseOrder(t, s, sup.ID, []model.PurchaseItem{{ProductID: p.ID, Quantity: 1, UnitPrice: 100}})

	// 未确认不可入库
	if _, err := s.CreateInboundOrder(po.ID); !model.IsValidationError(err) {
		t.Fatalf("expected ValidationError for draft PO, got %v", err)
	}
}

func TestFullInboundFlow(t *testing.T) {
	s := newTestService()
	_, _, po := setupConfirmedPO(t, s)
	inbound, _ := s.CreateInboundOrder(po.ID)

	// 开始质检
	ins, err := s.StartInspection(inbound.ID, "质检员小王")
	if err != nil {
		t.Fatalf("StartInspection: %v", err)
	}
	if ins.Result != model.InspectionPending {
		t.Fatalf("result = %q", ins.Result)
	}
	// 入库单状态应变为质检中
	inbound2, _ := s.GetInboundOrder(inbound.ID)
	if inbound2.Status != model.InboundInspecting {
		t.Fatalf("inbound status = %q, want inspecting", inbound2.Status)
	}

	// 质检合格
	if _, err := s.CompleteInspection(ins.ID, model.InspectionPassed, 0, "合格"); err != nil {
		t.Fatalf("CompleteInspection: %v", err)
	}

	// 入库
	stocked, err := s.StockInbound(inbound.ID)
	if err != nil {
		t.Fatalf("StockInbound: %v", err)
	}
	if stocked.Status != model.InboundStocked {
		t.Fatalf("inbound status = %q, want stocked", stocked.Status)
	}

	// 采购单应变为已收货
	poAfter, _ := s.GetPurchaseOrder(po.ID)
	if poAfter.Status != model.POStatusReceived {
		t.Fatalf("po status = %q, want received", poAfter.Status)
	}

	// 库存应增加
	stock, _ := s.GetStock(po.Items[0].ProductID)
	if stock != 10 {
		t.Fatalf("stock = %d, want 10", stock)
	}
}

func TestStockInboundRequiresPassedInspection(t *testing.T) {
	s := newTestService()
	_, _, po := setupConfirmedPO(t, s)
	inbound, _ := s.CreateInboundOrder(po.ID)
	ins, _ := s.StartInspection(inbound.ID, "质检员")

	// 质检未完成不可入库
	if _, err := s.StockInbound(inbound.ID); !model.IsValidationError(err) {
		t.Fatalf("expected ValidationError for pending inspection, got %v", err)
	}

	// 质检不合格不可入库
	s.CompleteInspection(ins.ID, model.InspectionFailed, 3, "不合格")
	if _, err := s.StockInbound(inbound.ID); !model.IsValidationError(err) {
		t.Fatalf("expected ValidationError for failed inspection, got %v", err)
	}
}

func TestRejectInbound(t *testing.T) {
	s := newTestService()
	_, _, po := setupConfirmedPO(t, s)
	inbound, _ := s.CreateInboundOrder(po.ID)

	rejected, err := s.RejectInbound(inbound.ID)
	if err != nil {
		t.Fatalf("RejectInbound: %v", err)
	}
	if rejected.Status != model.InboundRejected {
		t.Fatalf("status = %q", rejected.Status)
	}

	// 已驳回不可再驳回
	if _, err := s.RejectInbound(inbound.ID); !model.IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestInspectionLifecycle(t *testing.T) {
	s := newTestService()
	_, _, po := setupConfirmedPO(t, s)
	inbound, _ := s.CreateInboundOrder(po.ID)
	ins, _ := s.StartInspection(inbound.ID, "质检员")

	// 完成质检后不可重复流转
	if _, err := s.CompleteInspection(ins.ID, model.InspectionFailed, 1, "瑕疵"); err != nil {
		t.Fatalf("CompleteInspection: %v", err)
	}
	if _, err := s.CompleteInspection(ins.ID, model.InspectionPassed, 0, ""); !model.IsValidationError(err) {
		t.Fatalf("expected ValidationError re-completing, got %v", err)
	}
}
