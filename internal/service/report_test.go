package service

import (
	"testing"

	"supplychain/internal/model"
)

func TestStockReport(t *testing.T) {
	s := newTestService()
	p1 := seedProduct(t, s, "SKU001", "螺丝", 100, 10)
	p2 := seedProduct(t, s, "SKU002", "螺母", 200, 5)
	s.AdjustStock(p1.ID, 30, "入库")
	s.AdjustStock(p2.ID, 3, "入库")

	report, err := s.StockReport(false)
	if err != nil {
		t.Fatalf("StockReport: %v", err)
	}
	if len(report) != 2 {
		t.Fatalf("report len = %d, want 2", len(report))
	}
	// p2 库存 3 < 阈值 5，为低库存
	onlyLow, _ := s.StockReport(true)
	if len(onlyLow) != 1 || onlyLow[0].ProductID != p2.ID {
		t.Fatalf("onlyLow = %+v", onlyLow)
	}
	// 校验金额
	for _, item := range report {
		if item.ProductID == p1.ID && item.Value != 3000 {
			t.Fatalf("p1 value = %d, want 3000", item.Value)
		}
	}
}

func TestQualityStats(t *testing.T) {
	s := newTestService()
	sup := seedSupplier(t, s, "华东电子", "13800000001")
	p := seedProduct(t, s, "SKU001", "螺丝", 100, 10)
	po := seedPurchaseOrder(t, s, sup.ID, []model.PurchaseItem{{ProductID: p.ID, Quantity: 10, UnitPrice: 100}})
	s.ConfirmPurchaseOrder(po.ID)
	in1, _ := s.CreateInboundOrder(po.ID)
	in2, _ := s.CreateInboundOrder(po.ID) // 第二个会失败，忽略

	ins1, _ := s.StartInspection(in1.ID, "质检员")
	s.CompleteInspection(ins1.ID, model.InspectionPassed, 0, "合格")

	// 第二个采购单
	po2 := seedPurchaseOrder(t, s, sup.ID, []model.PurchaseItem{{ProductID: p.ID, Quantity: 5, UnitPrice: 100}})
	s.ConfirmPurchaseOrder(po2.ID)
	inb2, _ := s.CreateInboundOrder(po2.ID)
	ins2, _ := s.StartInspection(inb2.ID, "质检员")
	s.CompleteInspection(ins2.ID, model.InspectionFailed, 2, "不合格")

	stats, err := s.QualityStats()
	if err != nil {
		t.Fatalf("QualityStats: %v", err)
	}
	if stats.Total != 2 || stats.Passed != 1 || stats.Failed != 1 {
		t.Fatalf("stats = %+v", stats)
	}
	if stats.PassRate != 0.5 {
		t.Fatalf("pass rate = %v, want 0.5", stats.PassRate)
	}
	_ = in2
}

func TestSupplierRanking(t *testing.T) {
	s := newTestService()
	sup1 := seedSupplier(t, s, "供应商A", "13800000001")
	sup2 := seedSupplier(t, s, "供应商B", "13800000002")
	p := seedProduct(t, s, "SKU001", "螺丝", 100, 10)

	po1 := seedPurchaseOrder(t, s, sup1.ID, []model.PurchaseItem{{ProductID: p.ID, Quantity: 100, UnitPrice: 10}})
	s.ConfirmPurchaseOrder(po1.ID)
	po2 := seedPurchaseOrder(t, s, sup2.ID, []model.PurchaseItem{{ProductID: p.ID, Quantity: 10, UnitPrice: 10}})
	s.ConfirmPurchaseOrder(po2.ID)
	// 草稿单不计入排行
	seedPurchaseOrder(t, s, sup2.ID, []model.PurchaseItem{{ProductID: p.ID, Quantity: 999, UnitPrice: 100}})

	ranking, err := s.SupplierRanking(10)
	if err != nil {
		t.Fatalf("SupplierRanking: %v", err)
	}
	if len(ranking) != 2 {
		t.Fatalf("ranking len = %d, want 2", len(ranking))
	}
	if ranking[0].SupplierID != sup1.ID || ranking[0].TotalAmount != 1000 {
		t.Fatalf("ranking[0] = %+v", ranking[0])
	}
	if ranking[1].SupplierID != sup2.ID || ranking[1].TotalAmount != 100 {
		t.Fatalf("ranking[1] = %+v", ranking[1])
	}
}

func TestInventoryValue(t *testing.T) {
	s := newTestService()
	p1 := seedProduct(t, s, "SKU001", "螺丝", 100, 10)
	p2 := seedProduct(t, s, "SKU002", "螺母", 200, 5)
	s.AdjustStock(p1.ID, 10, "入库") // 1000
	s.AdjustStock(p2.ID, 5, "入库")  // 1000

	value, err := s.InventoryValue()
	if err != nil {
		t.Fatalf("InventoryValue: %v", err)
	}
	if value.TotalValue != 2000 || value.TotalStock != 2 || value.SKUCount != 2 {
		t.Fatalf("value = %+v", value)
	}
}
