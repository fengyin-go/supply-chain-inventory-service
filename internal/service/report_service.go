package service

import (
	"sort"

	"supplychain/internal/model"
)

// StockReportItem 库存报表项。
type StockReportItem struct {
	ProductID string `json:"product_id"`
	SKU       string `json:"sku"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	Stock     int    `json:"stock"`
	Price     int64  `json:"price"`
	Value     int64  `json:"value"`     // 库存金额 = 库存 × 单价
	LowStock  bool   `json:"low_stock"` // 是否低于阈值
}

// StockReport 库存报表（含低库存预警）。
func (s *Service) StockReport(onlyLowStock bool) ([]*StockReportItem, error) {
	products := s.store.ListProducts()
	report := make([]*StockReportItem, 0, len(products))
	for _, p := range products {
		stock, _ := s.GetStock(p.ID)
		item := &StockReportItem{
			ProductID: p.ID,
			SKU:       p.SKU,
			Name:      p.Name,
			Category:  p.Category,
			Stock:     stock,
			Price:     p.Price,
			Value:     int64(stock) * p.Price,
			LowStock:  stock < p.LowStockThreshold,
		}
		if onlyLowStock && !item.LowStock {
			continue
		}
		report = append(report, item)
	}
	sort.Slice(report, func(i, j int) bool {
		return report[i].Stock < report[j].Stock
	})
	return report, nil
}

// QualityStats 质检合格率统计。
type QualityStats struct {
	Total     int     `json:"total"`
	Passed    int     `json:"passed"`
	Failed    int     `json:"failed"`
	Pending   int     `json:"pending"`
	PassRate  float64 `json:"pass_rate"` // 合格率 = 合格 / (合格 + 不合格)
}

func (s *Service) QualityStats() (*QualityStats, error) {
	inspections := s.store.ListInspections()
	stats := &QualityStats{Total: len(inspections)}
	for _, i := range inspections {
		switch i.Result {
		case model.InspectionPassed:
			stats.Passed++
		case model.InspectionFailed:
			stats.Failed++
		case model.InspectionPending:
			stats.Pending++
		}
	}
	closed := stats.Passed + stats.Failed
	if closed > 0 {
		stats.PassRate = float64(stats.Passed) / float64(closed)
	}
	return stats, nil
}

// SupplierRankingItem 供应商采购额排行项。
type SupplierRankingItem struct {
	SupplierID   string `json:"supplier_id"`
	SupplierName string `json:"supplier_name"`
	OrderCount   int    `json:"order_count"`
	TotalAmount  int64  `json:"total_amount"` // 单位分
}

// SupplierRanking 供应商采购额排行（按已确认/已收货采购单金额）。
func (s *Service) SupplierRanking(limit int) ([]*SupplierRankingItem, error) {
	orders := s.store.ListPurchaseOrders()
	agg := map[string]*SupplierRankingItem{}
	for _, o := range orders {
		if o.Status == model.POStatusDraft || o.Status == model.POStatusCancelled {
			continue
		}
		item, ok := agg[o.SupplierID]
		if !ok {
			item = &SupplierRankingItem{SupplierID: o.SupplierID}
			if sup, err := s.store.GetSupplier(o.SupplierID); err == nil {
				item.SupplierName = sup.Name
			}
			agg[o.SupplierID] = item
		}
		item.OrderCount++
		item.TotalAmount += o.TotalAmount
	}
	ranking := make([]*SupplierRankingItem, 0, len(agg))
	for _, item := range agg {
		ranking = append(ranking, item)
	}
	sort.Slice(ranking, func(i, j int) bool {
		return ranking[i].TotalAmount > ranking[j].TotalAmount
	})
	if limit > 0 && len(ranking) > limit {
		ranking = ranking[:limit]
	}
	return ranking, nil
}

// InventoryValue 库存总金额统计。
type InventoryValue struct {
	TotalValue int64 `json:"total_value"` // 单位分
	TotalStock int   `json:"total_stock"` // 商品种类数
	SKUCount   int   `json:"sku_count"`
}

func (s *Service) InventoryValue() (*InventoryValue, error) {
	report, err := s.StockReport(false)
	if err != nil {
		return nil, err
	}
	value := &InventoryValue{SKUCount: len(report)}
	for _, item := range report {
		value.TotalValue += item.Value
		if item.Stock > 0 {
			value.TotalStock++
		}
	}
	return value, nil
}

// PurchaseSummaryItem 采购汇总项。
type PurchaseSummaryItem struct {
	SupplierID   string `json:"supplier_id"`
	SupplierName string `json:"supplier_name"`
	OrderCount   int    `json:"order_count"`
	TotalAmount  int64  `json:"total_amount"` // 单位分
	TotalQuantity int   `json:"total_quantity"`
}

// PurchaseSummary 采购汇总（按供应商，统计已确认/已收货采购单）。
func (s *Service) PurchaseSummary() ([]*PurchaseSummaryItem, error) {
	orders := s.store.ListPurchaseOrders()
	agg := map[string]*PurchaseSummaryItem{}
	for _, o := range orders {
		if o.Status == model.POStatusDraft || o.Status == model.POStatusCancelled {
			continue
		}
		item, ok := agg[o.SupplierID]
		if !ok {
			item = &PurchaseSummaryItem{SupplierID: o.SupplierID}
			if sup, err := s.store.GetSupplier(o.SupplierID); err == nil {
				item.SupplierName = sup.Name
			}
			agg[o.SupplierID] = item
		}
		item.OrderCount++
		item.TotalAmount += o.TotalAmount
		for _, it := range o.Items {
			item.TotalQuantity += it.Quantity
		}
	}
	summary := make([]*PurchaseSummaryItem, 0, len(agg))
	for _, item := range agg {
		summary = append(summary, item)
	}
	sort.Slice(summary, func(i, j int) bool {
		return summary[i].TotalAmount > summary[j].TotalAmount
	})
	return summary, nil
}

// MovementStats 库存流水统计。
type MovementStats struct {
	InboundCount  int `json:"inbound_count"`
	OutboundCount int `json:"outbound_count"`
	AdjustCount   int `json:"adjust_count"`
	InboundQty    int `json:"inbound_qty"`
	OutboundQty   int `json:"outbound_qty"`
}

func (s *Service) MovementStats() (*MovementStats, error) {
	movements := s.store.ListMovements()
	stats := &MovementStats{}
	for _, m := range movements {
		switch m.Type {
		case model.MovementInbound:
			stats.InboundCount++
			stats.InboundQty += m.Delta
		case model.MovementOutbound:
			stats.OutboundCount++
			stats.OutboundQty += -m.Delta
		case model.MovementAdjust:
			stats.AdjustCount++
		}
	}
	return stats, nil
}

// CategoryStatsItem 商品分类统计项。
type CategoryStatsItem struct {
	Category   string `json:"category"`
	SKUCount   int    `json:"sku_count"`
	Stock      int    `json:"stock"`
	StockValue int64  `json:"stock_value"` // 单位分
}

// CategoryStats 按分类统计商品与库存。
func (s *Service) CategoryStats() ([]*CategoryStatsItem, error) {
	report, err := s.StockReport(false)
	if err != nil {
		return nil, err
	}
	agg := map[string]*CategoryStatsItem{}
	for _, item := range report {
		stat, ok := agg[item.Category]
		if !ok {
			stat = &CategoryStatsItem{Category: item.Category}
			agg[item.Category] = stat
		}
		stat.SKUCount++
		stat.Stock += item.Stock
		stat.StockValue += item.Value
	}
	stats := make([]*CategoryStatsItem, 0, len(agg))
	for _, stat := range agg {
		stats = append(stats, stat)
	}
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].StockValue > stats[j].StockValue
	})
	return stats, nil
}

// BatchTrace 批次追溯信息。
type BatchTrace struct {
	BatchNo           string `json:"batch_no"`
	ProductID         string `json:"product_id"`
	ProductName       string `json:"product_name"`
	Quantity          int    `json:"quantity"`
	Remaining         int    `json:"remaining"`
	InboundNo         string `json:"inbound_no"`
	PurchaseOrderNo   string `json:"purchase_order_no"`
	SupplierName      string `json:"supplier_name"`
	InspectionResult  string `json:"inspection_result"`
	Inspector         string `json:"inspector"`
}

// TraceBatch 追溯某库存批次的来源链路（入库单 → 采购单 → 供应商 → 质检）。
func (s *Service) TraceBatch(batchID string) (*BatchTrace, error) {
	batch, err := s.store.GetBatch(batchID)
	if err != nil {
		return nil, err
	}
	trace := &BatchTrace{
		BatchNo:   batch.BatchNo,
		ProductID: batch.ProductID,
		Quantity:  batch.Quantity,
		Remaining: batch.Remaining,
	}
	if p, err := s.store.GetProduct(batch.ProductID); err == nil {
		trace.ProductName = p.Name
	}
	inbound, err := s.store.GetInboundOrder(batch.InboundOrderID)
	if err == nil {
		trace.InboundNo = inbound.InboundNo
		if po, err := s.store.GetPurchaseOrder(inbound.PurchaseOrderID); err == nil {
			trace.PurchaseOrderNo = po.OrderNo
		}
		if sup, err := s.store.GetSupplier(inbound.SupplierID); err == nil {
			trace.SupplierName = sup.Name
		}
		if ins, err := s.store.GetInspectionByInbound(inbound.ID); err == nil {
			trace.InspectionResult = ins.Result
			trace.Inspector = ins.Inspector
		}
	}
	return trace, nil
}
