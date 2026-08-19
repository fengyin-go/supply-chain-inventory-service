package model

import (
	"strings"
	"time"
)

// 库存流水类型。
const (
	MovementInbound  = "inbound"  // 入库
	MovementOutbound = "outbound" // 出库
	MovementAdjust   = "adjust"   // 盘点调整
)

// StockMovement 库存流水。
type StockMovement struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	BatchID   string    `json:"batch_id,omitempty"`
	Type      string    `json:"type"`
	Delta     int       `json:"delta"` // 正数增加、负数减少
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

func (m *StockMovement) Validate() error {
	m.ProductID = strings.TrimSpace(m.ProductID)
	m.Type = strings.TrimSpace(m.Type)
	m.Note = strings.TrimSpace(m.Note)
	if m.ProductID == "" {
		return NewValidationError("product_id", "商品 ID 不能为空")
	}
	if m.Delta == 0 {
		return NewValidationError("delta", "变动数量不能为 0")
	}
	if m.Type != MovementInbound && m.Type != MovementOutbound && m.Type != MovementAdjust {
		return NewValidationError("type", "流水类型不合法")
	}
	return nil
}

// MovementFilter 库存流水筛选条件。
type MovementFilter struct {
	ProductID string
	Type      string
}

func (f MovementFilter) Match(m *StockMovement) bool {
	if f.ProductID != "" && m.ProductID != f.ProductID {
		return false
	}
	if f.Type != "" && m.Type != f.Type {
		return false
	}
	return true
}
