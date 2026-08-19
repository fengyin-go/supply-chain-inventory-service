package service

import (
	"testing"

	"supplychain/internal/model"
)

func TestRejectInboundAfterInspection(t *testing.T) {
	s := newTestService()
	_, _, po := setupConfirmedPO(t, s)
	inbound, _ := s.CreateInboundOrder(po.ID)
	s.StartInspection(inbound.ID, "质检员")

	// 质检中状态可驳回
	rejected, err := s.RejectInbound(inbound.ID)
	if err != nil {
		t.Fatalf("RejectInbound: %v", err)
	}
	if rejected.Status != model.InboundRejected {
		t.Fatalf("status = %q", rejected.Status)
	}
}

func TestCompleteInspectionNegativeDefect(t *testing.T) {
	s := newTestService()
	_, _, po := setupConfirmedPO(t, s)
	inbound, _ := s.CreateInboundOrder(po.ID)
	ins, _ := s.StartInspection(inbound.ID, "质检员")

	// 负缺陷数量 → 校验失败
	if _, err := s.CompleteInspection(ins.ID, model.InspectionPassed, -1, ""); !model.IsValidationError(err) {
		t.Fatalf("expected ValidationError for negative defect, got %v", err)
	}
}

func TestCreateInboundDuplicate(t *testing.T) {
	s := newTestService()
	_, _, po := setupConfirmedPO(t, s)
	if _, err := s.CreateInboundOrder(po.ID); err != nil {
		t.Fatalf("first inbound: %v", err)
	}
	// 第二次创建同一采购单入库 → 校验失败
	if _, err := s.CreateInboundOrder(po.ID); !model.IsValidationError(err) {
		t.Fatalf("expected ValidationError for duplicate inbound, got %v", err)
	}
}

func TestConfirmNonDraftPO(t *testing.T) {
	s := newTestService()
	sup := seedSupplier(t, s, "华东电子", "13800000001")
	p := seedProduct(t, s, "SKU001", "螺丝", 100, 10)
	po := seedPurchaseOrder(t, s, sup.ID, []model.PurchaseItem{{ProductID: p.ID, Quantity: 1, UnitPrice: 100}})
	s.CancelPurchaseOrder(po.ID)

	// 已取消不可确认
	if _, err := s.ConfirmPurchaseOrder(po.ID); !model.IsValidationError(err) {
		t.Fatalf("expected ValidationError confirming cancelled, got %v", err)
	}
}

func TestSupplierRankingExcludesCancelled(t *testing.T) {
	s := newTestService()
	sup := seedSupplier(t, s, "华东电子", "13800000001")
	p := seedProduct(t, s, "SKU001", "螺丝", 100, 10)

	// 草稿与取消单不计入
	seedPurchaseOrder(t, s, sup.ID, []model.PurchaseItem{{ProductID: p.ID, Quantity: 100, UnitPrice: 100}})
	po := seedPurchaseOrder(t, s, sup.ID, []model.PurchaseItem{{ProductID: p.ID, Quantity: 50, UnitPrice: 100}})
	s.CancelPurchaseOrder(po.ID)

	ranking, err := s.SupplierRanking(10)
	if err != nil {
		t.Fatalf("SupplierRanking: %v", err)
	}
	if len(ranking) != 0 {
		t.Fatalf("ranking = %+v, want empty (no confirmed orders)", ranking)
	}
}

func TestListReturnOrdersEmpty(t *testing.T) {
	s := newTestService()
	items, total, err := s.ListReturnOrders(model.ReturnFilter{}, 1, 10)
	if err != nil {
		t.Fatalf("ListReturnOrders: %v", err)
	}
	if total != 0 || len(items) != 0 {
		t.Fatalf("empty list total=%d len=%d", total, len(items))
	}
}
