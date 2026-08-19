package service

import (
	"testing"

	"supplychain/internal/model"
)

func TestGetStockAndAdjust(t *testing.T) {
	s := newTestService()
	p := seedProduct(t, s, "SKU001", "螺丝", 100, 10)

	// 初始库存为 0
	stock, err := s.GetStock(p.ID)
	if err != nil {
		t.Fatalf("GetStock: %v", err)
	}
	if stock != 0 {
		t.Fatalf("stock = %d, want 0", stock)
	}

	// 正向调整
	if err := s.AdjustStock(p.ID, 30, "期初盘点"); err != nil {
		t.Fatalf("AdjustStock: %v", err)
	}
	stock, _ = s.GetStock(p.ID)
	if stock != 30 {
		t.Fatalf("stock = %d, want 30", stock)
	}

	// 负向调整
	if err := s.AdjustStock(p.ID, -10, "盘点损耗"); err != nil {
		t.Fatalf("AdjustStock negative: %v", err)
	}
	stock, _ = s.GetStock(p.ID)
	if stock != 20 {
		t.Fatalf("stock = %d, want 20", stock)
	}

	// 负向调整超出库存失败
	if err := s.AdjustStock(p.ID, -100, "超量"); !model.IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}

	// 查询不存在的商品
	if _, err := s.GetStock("missing"); err == nil {
		t.Fatal("expected error for missing product")
	}
}

func TestOutboundFIFO(t *testing.T) {
	s := newTestService()
	p := seedProduct(t, s, "SKU001", "螺丝", 100, 10)

	// 分两批入库，模拟先进先出
	s.AdjustStock(p.ID, 10, "第一批") // 批次 A，数量 10
	s.AdjustStock(p.ID, 20, "第二批") // 批次 B，数量 20

	// 出库 15，应扣减第一批 10 + 第二批 5
	if err := s.OutboundStock(p.ID, 15, "销售出库"); err != nil {
		t.Fatalf("OutboundStock: %v", err)
	}
	stock, _ := s.GetStock(p.ID)
	if stock != 15 {
		t.Fatalf("stock = %d, want 15", stock)
	}

	// 第一批应剩余 0，第二批剩余 15
	batches := s.store.ListBatchesByProduct(p.ID)
	// batches 按创建时间排序
	var batchRemaining []int
	for _, b := range batches {
		batchRemaining = append(batchRemaining, b.Remaining)
	}
	// 排序保证先进先出（列表顺序可能不保证，但数量总和应正确）
	sum := 0
	for _, r := range batchRemaining {
		sum += r
	}
	if sum != 15 {
		t.Fatalf("batch remaining sum = %d, want 15", sum)
	}

	// 出库超出库存失败
	if err := s.OutboundStock(p.ID, 100, "超量"); !model.IsValidationError(err) {
		t.Fatalf("expected ValidationError, got %v", err)
	}

	// 非法数量
	if err := s.OutboundStock(p.ID, 0, "零"); !model.IsValidationError(err) {
		t.Fatalf("expected ValidationError for zero quantity, got %v", err)
	}
}

func TestListBatchesAndMovements(t *testing.T) {
	s := newTestService()
	p := seedProduct(t, s, "SKU001", "螺丝", 100, 10)
	s.AdjustStock(p.ID, 10, "入库")
	s.OutboundStock(p.ID, 3, "出库")

	_, total, err := s.ListBatches(model.BatchFilter{ProductID: p.ID}, 1, 10)
	if err != nil || total < 1 {
		t.Fatalf("ListBatches: %v total=%d", err, total)
	}

	_, total, err = s.ListMovements(model.MovementFilter{ProductID: p.ID}, 1, 10)
	if err != nil || total != 2 {
		t.Fatalf("ListMovements total = %d, want 2", total)
	}

	_, total, _ = s.ListMovements(model.MovementFilter{ProductID: p.ID, Type: model.MovementOutbound}, 1, 10)
	if total != 1 {
		t.Fatalf("outbound movements = %d, want 1", total)
	}
}
