package service

import (
	"testing"

	"supplychain/internal/model"
)

// TestMultiProductFullFlow 覆盖多商品采购 → 入库 → 质检 → 库存 → 出库的完整闭环。
func TestMultiProductFullFlow(t *testing.T) {
	s := newTestService()
	sup := seedSupplier(t, s, "华东电子", "13800000001")
	p1 := seedProduct(t, s, "SKU001", "螺丝", 100, 10)
	p2 := seedProduct(t, s, "SKU002", "螺母", 200, 5)

	po := seedPurchaseOrder(t, s, sup.ID, []model.PurchaseItem{
		{ProductID: p1.ID, Quantity: 20, UnitPrice: 100},
		{ProductID: p2.ID, Quantity: 10, UnitPrice: 200},
	})
	s.ConfirmPurchaseOrder(po.ID)
	inbound, _ := s.CreateInboundOrder(po.ID)
	ins, _ := s.StartInspection(inbound.ID, "质检员")
	s.CompleteInspection(ins.ID, model.InspectionPassed, 0, "合格")
	s.StockInbound(inbound.ID)

	// 两个商品分别有库存
	stock1, _ := s.GetStock(p1.ID)
	stock2, _ := s.GetStock(p2.ID)
	if stock1 != 20 || stock2 != 10 {
		t.Fatalf("stock = %d/%d, want 20/10", stock1, stock2)
	}

	// 出库 p1
	s.OutboundStock(p1.ID, 5, "出库p1")
	stock1, _ = s.GetStock(p1.ID)
	if stock1 != 15 {
		t.Fatalf("p1 stock = %d, want 15", stock1)
	}

	// 报表校验
	report, _ := s.StockReport(false)
	if len(report) != 2 {
		t.Fatalf("report len = %d", len(report))
	}

	// 供应商排行金额 = 20*100 + 10*200 = 4000
	ranking, _ := s.SupplierRanking(10)
	if ranking[0].TotalAmount != 4000 {
		t.Fatalf("ranking amount = %d, want 4000", ranking[0].TotalAmount)
	}
}

func TestPurchaseOrderTotalAmountEdge(t *testing.T) {
	s := newTestService()
	sup := seedSupplier(t, s, "华东电子", "13800000001")
	p := seedProduct(t, s, "SKU001", "螺丝", 100, 10)

	// 单价为 0 时行金额为 0（允许，因为单价可为 0 的赠品）
	po, err := s.CreatePurchaseOrder(model.PurchaseOrder{
		SupplierID: sup.ID,
		Items:      []model.PurchaseItem{{ProductID: p.ID, Quantity: 10, UnitPrice: 0}},
	})
	if err != nil {
		t.Fatalf("CreatePurchaseOrder: %v", err)
	}
	if po.TotalAmount != 0 {
		t.Fatalf("total = %d, want 0", po.TotalAmount)
	}
}

func TestSupplierProductListPagination(t *testing.T) {
	s := newTestService()
	for i := 0; i < 3; i++ {
		seedSupplier(t, s, "供应商"+string(rune('A'+i)), "1380000000"+string(rune('1'+i)))
		seedProduct(t, s, "SKU00"+string(rune('1'+i)), "商品"+string(rune('A'+i)), 100, 10)
	}

	// 供应商分页
	items, total, _ := s.ListSuppliers(model.SupplierFilter{}, 2, 1)
	if total != 3 || len(items) != 1 {
		t.Fatalf("supplier page total=%d len=%d", total, len(items))
	}

	// 商品分页越界
	empty, _, _ := s.ListProducts(model.ProductFilter{}, 99, 1)
	if len(empty) != 0 {
		t.Fatalf("empty page len = %d", len(empty))
	}
}

func TestGetStockValidatesProduct(t *testing.T) {
	s := newTestService()
	if _, err := s.GetStock("missing"); err == nil {
		t.Fatal("expected error for missing product")
	}
}

func TestGenerateNoFormat(t *testing.T) {
	no := generateNo("IN")
	if len(no) < 10 {
		t.Fatalf("generated no too short: %s", no)
	}
	if no[:2] != "IN" {
		t.Fatalf("prefix = %q", no[:2])
	}
}
