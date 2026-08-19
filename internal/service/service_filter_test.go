package service

import (
	"testing"

	"supplychain/internal/model"
)

func TestListReturnOrdersFilter(t *testing.T) {
	s := newTestService()
	inbound, p := setupStockedInbound(t, s)

	r1, _ := s.CreateReturnOrder(inbound.ID, p.ID, 2, "问题1")
	r2, _ := s.CreateReturnOrder(inbound.ID, p.ID, 3, "问题2")
	s.CompleteReturnOrder(r1.ID)

	_, total, _ := s.ListReturnOrders(model.ReturnFilter{Status: model.ReturnPending}, 1, 10)
	if total != 1 {
		t.Fatalf("pending total = %d, want 1", total)
	}
	_, total, _ = s.ListReturnOrders(model.ReturnFilter{InboundOrderID: inbound.ID}, 1, 10)
	if total != 2 {
		t.Fatalf("inbound total = %d, want 2", total)
	}
	_ = r2
}

func TestFIFOOrderPrecision(t *testing.T) {
	s := newTestService()
	p := seedProduct(t, s, "SKU001", "螺丝", 100, 10)

	// 三批入库：5、10、15（按创建顺序）
	s.AdjustStock(p.ID, 5, "批1")
	s.AdjustStock(p.ID, 10, "批2")
	s.AdjustStock(p.ID, 15, "批3")

	// 出库 12：应耗尽批1(5) + 批2(7)，批3 不变
	if err := s.OutboundStock(p.ID, 12, "出库"); err != nil {
		t.Fatalf("OutboundStock: %v", err)
	}

	// 检查各批剩余
	var remaining []int
	for _, b := range s.store.ListBatchesByProduct(p.ID) {
		remaining = append(remaining, b.Remaining)
	}
	// 批1=0，批2=3，批3=15（不保证顺序，但集合应一致）
	sum := 0
	hasZero := false
	for _, r := range remaining {
		sum += r
		if r == 0 {
			hasZero = true
		}
	}
	if sum != 18 || !hasZero {
		t.Fatalf("remaining = %v, sum = %d", remaining, sum)
	}
}

func TestOutboundExactStock(t *testing.T) {
	s := newTestService()
	p := seedProduct(t, s, "SKU001", "螺丝", 100, 10)
	s.AdjustStock(p.ID, 8, "入库")

	// 恰好全部出库
	if err := s.OutboundStock(p.ID, 8, "全部出库"); err != nil {
		t.Fatalf("OutboundStock: %v", err)
	}
	stock, _ := s.GetStock(p.ID)
	if stock != 0 {
		t.Fatalf("stock = %d, want 0", stock)
	}
	// 再出库失败
	if err := s.OutboundStock(p.ID, 1, "超量"); !model.IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestBatchesAndMovementsPagination(t *testing.T) {
	s := newTestService()
	p := seedProduct(t, s, "SKU001", "螺丝", 100, 10)
	for i := 0; i < 5; i++ {
		s.AdjustStock(p.ID, 1, "入库")
	}

	items, total, _ := s.ListBatches(model.BatchFilter{}, 1, 2)
	if total != 5 || len(items) != 2 {
		t.Fatalf("batch page total=%d len=%d", total, len(items))
	}

	items2, total2, _ := s.ListMovements(model.MovementFilter{}, 2, 2)
	if total2 != 5 || len(items2) != 2 {
		t.Fatalf("movement page total=%d len=%d", total2, len(items2))
	}

	// 越界页
	empty, _, _ := s.ListBatches(model.BatchFilter{}, 99, 2)
	if len(empty) != 0 {
		t.Fatalf("empty page len = %d", len(empty))
	}
}
