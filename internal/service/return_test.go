package service

import (
	"testing"

	"supplychain/internal/model"
)

// setupStockedInbound 建立一条已入库的完整链路，返回入库单与商品。
func setupStockedInbound(t *testing.T, s *Service) (*model.InboundOrder, *model.Product) {
	t.Helper()
	sup := seedSupplier(t, s, "华东电子", "13800000001")
	p := seedProduct(t, s, "SKU001", "螺丝", 100, 10)
	po := seedPurchaseOrder(t, s, sup.ID, []model.PurchaseItem{{ProductID: p.ID, Quantity: 20, UnitPrice: 100}})
	s.ConfirmPurchaseOrder(po.ID)
	inbound, _ := s.CreateInboundOrder(po.ID)
	ins, _ := s.StartInspection(inbound.ID, "质检员")
	s.CompleteInspection(ins.ID, model.InspectionPassed, 0, "合格")
	s.StockInbound(inbound.ID)
	return inbound, p
}

func TestReturnOrderLifecycle(t *testing.T) {
	s := newTestService()
	inbound, p := setupStockedInbound(t, s)

	ret, err := s.CreateReturnOrder(inbound.ID, p.ID, 5, "质量问题")
	if err != nil {
		t.Fatalf("CreateReturnOrder: %v", err)
	}
	if ret.Status != model.ReturnPending {
		t.Fatalf("status = %q", ret.Status)
	}
	if ret.ReturnNo == "" {
		t.Fatal("expected return no")
	}

	// 完成退货，库存应减少
	completed, err := s.CompleteReturnOrder(ret.ID)
	if err != nil {
		t.Fatalf("CompleteReturnOrder: %v", err)
	}
	if completed.Status != model.ReturnCompleted {
		t.Fatalf("status = %q", completed.Status)
	}
	stock, _ := s.GetStock(p.ID)
	if stock != 15 {
		t.Fatalf("stock = %d, want 15", stock)
	}

	// 重复完成失败
	if _, err := s.CompleteReturnOrder(ret.ID); !model.IsValidationError(err) {
		t.Fatalf("expected ValidationError re-completing, got %v", err)
	}
}

func TestReturnOrderValidation(t *testing.T) {
	s := newTestService()
	inbound, p := setupStockedInbound(t, s)

	// 商品不属于该入库单
	other := seedProduct(t, s, "SKU002", "螺母", 200, 5)
	if _, err := s.CreateReturnOrder(inbound.ID, other.ID, 1, "x"); !model.IsValidationError(err) {
		t.Fatalf("expected ValidationError for foreign product, got %v", err)
	}

	// 数量超过入库数量
	if _, err := s.CreateReturnOrder(inbound.ID, p.ID, 999, "x"); !model.IsValidationError(err) {
		t.Fatalf("expected ValidationError for over quantity, got %v", err)
	}

	// 未入库的入库单不可退货
	sup := seedSupplier(t, s, "供应商B", "13800000002")
	po := seedPurchaseOrder(t, s, sup.ID, []model.PurchaseItem{{ProductID: p.ID, Quantity: 5, UnitPrice: 100}})
	s.ConfirmPurchaseOrder(po.ID)
	inbound2, _ := s.CreateInboundOrder(po.ID) // 未入库
	if _, err := s.CreateReturnOrder(inbound2.ID, p.ID, 1, "x"); !model.IsValidationError(err) {
		t.Fatalf("expected ValidationError for non-stocked inbound, got %v", err)
	}
}

func TestReturnStats(t *testing.T) {
	s := newTestService()
	inbound, p := setupStockedInbound(t, s)

	r1, _ := s.CreateReturnOrder(inbound.ID, p.ID, 3, "问题1")
	s.CreateReturnOrder(inbound.ID, p.ID, 2, "问题2")
	s.CompleteReturnOrder(r1.ID)

	stats, err := s.ReturnStats()
	if err != nil {
		t.Fatalf("ReturnStats: %v", err)
	}
	if stats.Total != 2 || stats.Completed != 1 || stats.Pending != 1 || stats.TotalQuantity != 5 {
		t.Fatalf("stats = %+v", stats)
	}
}

func TestTraceBatch(t *testing.T) {
	s := newTestService()
	inbound, _ := setupStockedInbound(t, s)

	batches := s.store.ListBatches()
	var target string
	for _, b := range batches {
		if b.InboundOrderID == inbound.ID {
			target = b.ID
		}
	}
	if target == "" {
		t.Fatal("expected a batch for the inbound order")
	}
	trace, err := s.TraceBatch(target)
	if err != nil {
		t.Fatalf("TraceBatch: %v", err)
	}
	if trace.InboundNo == "" || trace.PurchaseOrderNo == "" || trace.SupplierName == "" {
		t.Fatalf("trace = %+v", trace)
	}
	if trace.InspectionResult != model.InspectionPassed {
		t.Fatalf("inspection = %q", trace.InspectionResult)
	}
}
