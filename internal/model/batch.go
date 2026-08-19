package model

import (
	"strings"
	"time"
)

// InventoryBatch 库存批次。入库质检通过后生成，记录一批商品的剩余量。
type InventoryBatch struct {
	ID             string    `json:"id"`
	BatchNo        string    `json:"batch_no"`
	ProductID      string    `json:"product_id"`
	InboundOrderID string    `json:"inbound_order_id"`
	Quantity       int       `json:"quantity"`  // 初始入库数量
	Remaining      int       `json:"remaining"` // 剩余数量
	CreatedAt      time.Time `json:"created_at"`
}

func (b *InventoryBatch) Validate() error {
	b.BatchNo = strings.TrimSpace(b.BatchNo)
	b.ProductID = strings.TrimSpace(b.ProductID)
	b.InboundOrderID = strings.TrimSpace(b.InboundOrderID)
	if b.BatchNo == "" {
		return NewValidationError("batch_no", "批次号不能为空")
	}
	if b.ProductID == "" {
		return NewValidationError("product_id", "商品 ID 不能为空")
	}
	if b.Quantity <= 0 {
		return NewValidationError("quantity", "批次数量必须大于 0")
	}
	if b.Remaining < 0 {
		return NewValidationError("remaining", "剩余数量不能为负数")
	}
	return nil
}

func (b *InventoryBatch) Restore(quantity int) {
	b.Remaining += quantity
}

// BatchFilter 批次筛选条件。
type BatchFilter struct {
	ProductID string
}

func (f BatchFilter) Match(b *InventoryBatch) bool {
	if f.ProductID != "" && b.ProductID != f.ProductID {
		return false
	}
	return true
}
