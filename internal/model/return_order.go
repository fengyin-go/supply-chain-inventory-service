package model

import (
	"strings"
	"time"
)

// 退货单状态。
const (
	ReturnPending   = "pending"   // 待处理
	ReturnCompleted = "completed" // 已完成
)

// returnTransitions 退货单状态机。
var returnTransitions = map[string]map[string]bool{
	ReturnPending: {ReturnCompleted: true},
}

// CanTransitionReturn 判断退货单状态是否可流转。
func CanTransitionReturn(from, to string) bool {
	if m, ok := returnTransitions[from]; ok {
		return m[to]
	}
	return false
}

// ReturnOrder 采购退货单。
type ReturnOrder struct {
	ID             string    `json:"id"`
	ReturnNo       string    `json:"return_no"`
	InboundOrderID string    `json:"inbound_order_id"`
	ProductID      string    `json:"product_id"`
	Quantity       int       `json:"quantity"`
	Reason         string    `json:"reason"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (r *ReturnOrder) Validate() error {
	r.ReturnNo = strings.TrimSpace(r.ReturnNo)
	r.InboundOrderID = strings.TrimSpace(r.InboundOrderID)
	r.ProductID = strings.TrimSpace(r.ProductID)
	r.Reason = strings.TrimSpace(r.Reason)
	if r.InboundOrderID == "" {
		return NewValidationError("inbound_order_id", "入库单 ID 不能为空")
	}
	if r.ProductID == "" {
		return NewValidationError("product_id", "商品 ID 不能为空")
	}
	if r.Quantity <= 0 {
		return NewValidationError("quantity", "退货数量必须大于 0")
	}
	if r.Status == "" {
		r.Status = ReturnPending
	}
	if r.Status != ReturnPending && r.Status != ReturnCompleted {
		return NewValidationError("status", "退货单状态不合法")
	}
	return nil
}

// ReturnFilter 退货单筛选条件。
type ReturnFilter struct {
	Status         string
	InboundOrderID string
}

func (f ReturnFilter) Match(r *ReturnOrder) bool {
	if f.Status != "" && r.Status != f.Status {
		return false
	}
	if f.InboundOrderID != "" && r.InboundOrderID != f.InboundOrderID {
		return false
	}
	return true
}
