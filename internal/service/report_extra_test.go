package service

import (
	"testing"

	"supplychain/internal/model"
)

func TestPurchaseSummary(t *testing.T) {
	s := newTestService()
	sup1 := seedSupplier(t, s, "供应商A", "13800000001")
	sup2 := seedSupplier(t, s, "供应商B", "13800000002")
	p := seedProduct(t, s, "SKU001", "螺丝", 100, 10)

	po1 := seedPurchaseOrder(t, s, sup1.ID, []model.PurchaseItem{{ProductID: p.ID, Quantity: 100, UnitPrice: 10}})
	s.ConfirmPurchaseOrder(po1.ID)
	po2 := seedPurchaseOrder(t, s, sup2.ID, []model.PurchaseItem{{ProductID: p.ID, Quantity: 10, UnitPrice: 10}})
	s.ConfirmPurchaseOrder(po2.ID)

	summary, err := s.PurchaseSummary()
	if err != nil {
		t.Fatalf("PurchaseSummary: %v", err)
	}
	if len(summary) != 2 {
		t.Fatalf("len = %d, want 2", len(summary))
	}
	if summary[0].SupplierID != sup1.ID || summary[0].TotalQuantity != 100 {
		t.Fatalf("summary[0] = %+v", summary[0])
	}
}

func TestMovementStats(t *testing.T) {
	s := newTestService()
	sup := seedSupplier(t, s, "华东电子", "13800000001")
	p := seedProduct(t, s, "SKU001", "螺丝", 100, 10)
	po := seedPurchaseOrder(t, s, sup.ID, []model.PurchaseItem{{ProductID: p.ID, Quantity: 20, UnitPrice: 100}})
	s.ConfirmPurchaseOrder(po.ID)
	inbound, _ := s.CreateInboundOrder(po.ID)
	ins, _ := s.StartInspection(inbound.ID, "质检员")
	s.CompleteInspection(ins.ID, model.InspectionPassed, 0, "合格")
	s.StockInbound(inbound.ID) // 入库流水 +20

	s.OutboundStock(p.ID, 5, "出库") // 出库流水 -5
	s.AdjustStock(p.ID, -2, "损耗")   // 调整流水 -2

	stats, err := s.MovementStats()
	if err != nil {
		t.Fatalf("MovementStats: %v", err)
	}
	if stats.InboundCount != 1 || stats.OutboundCount != 1 || stats.AdjustCount != 1 {
		t.Fatalf("stats = %+v", stats)
	}
	if stats.InboundQty != 20 || stats.OutboundQty != 5 {
		t.Fatalf("qty stats = %+v", stats)
	}
}

func TestCategoryStats(t *testing.T) {
	s := newTestService()
	p1 := seedProduct(t, s, "SKU001", "螺丝", 100, 10)
	p2 := seedProduct(t, s, "SKU002", "螺母", 200, 5)
	// 设置不同分类
	s.UpdateProduct(p1.ID, model.Product{SKU: "SKU001", Name: "螺丝", Category: "紧固件", Price: 100, LowStockThreshold: 10})
	s.UpdateProduct(p2.ID, model.Product{SKU: "SKU002", Name: "螺母", Category: "五金", Price: 200, LowStockThreshold: 5})
	s.AdjustStock(p1.ID, 10, "入库")

	stats, err := s.CategoryStats()
	if err != nil {
		t.Fatalf("CategoryStats: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("len = %d, want 2", len(stats))
	}
	// 紧固件有库存，价值 1000
	for _, item := range stats {
		if item.Category == "紧固件" && item.StockValue != 1000 {
			t.Fatalf("紧固件 value = %d", item.StockValue)
		}
	}
}

func TestSupplierRankingLimit(t *testing.T) {
	s := newTestService()
	p := seedProduct(t, s, "SKU001", "螺丝", 100, 10)
	for i := 0; i < 5; i++ {
		sup := seedSupplier(t, s, "供应商"+string(rune('A'+i)), "1380000000"+string(rune('1'+i)))
		po := seedPurchaseOrder(t, s, sup.ID, []model.PurchaseItem{{ProductID: p.ID, Quantity: 1, UnitPrice: 100}})
		s.ConfirmPurchaseOrder(po.ID)
	}

	ranking, err := s.SupplierRanking(3)
	if err != nil {
		t.Fatalf("SupplierRanking: %v", err)
	}
	if len(ranking) != 3 {
		t.Fatalf("len = %d, want 3", len(ranking))
	}
}
