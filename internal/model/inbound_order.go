package model

import (
	"strings"
	"time"
)

// 入库单状态。
const (
	InboundPending    = "pending"    // 待入库
	InboundInspecting = "inspecting" // 质检中
	InboundStocked    = "stocked"    // 已入库
	InboundRejected   = "rejected"   // 已驳回
)

// inboundTransitions 入库单状态机。
var inboundTransitions = map[string]map[string]bool{
	InboundPending:    {InboundInspecting: true, InboundRejected: true},
	InboundInspecting: {InboundStocked: true, InboundRejected: true},
}

// CanTransitionInbound 判断入库单状态是否可流转。
func CanTransitionInbound(from, to string) bool {
	if m, ok := inboundTransitions[from]; ok {
		return m[to]
	}
	return false
}

// InboundOrder 入库单。
type InboundOrder struct {
	ID              string         `json:"id"`
	InboundNo       string         `json:"inbound_no"`
	PurchaseOrderID string         `json:"purchase_order_id"`
	SupplierID      string         `json:"supplier_id"`
	Items           []PurchaseItem `json:"items"`
	TotalAmount     int64          `json:"total_amount"`
	Status          string         `json:"status"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

func (i *InboundOrder) Validate() error {
	i.InboundNo = strings.TrimSpace(i.InboundNo)
	i.PurchaseOrderID = strings.TrimSpace(i.PurchaseOrderID)
	i.SupplierID = strings.TrimSpace(i.SupplierID)
	if i.PurchaseOrderID == "" {
		return NewValidationError("purchase_order_id", "采购单 ID 不能为空")
	}
	if len(i.Items) == 0 {
		return NewValidationError("items", "入库明细不能为空")
	}
	var total int64
	for idx := range i.Items {
		item := &i.Items[idx]
		item.ProductID = strings.TrimSpace(item.ProductID)
		if item.ProductID == "" {
			return NewValidationError("items", "明细中的商品 ID 不能为空")
		}
		if item.Quantity <= 0 {
			return NewValidationError("items", "明细中的数量必须大于 0")
		}
		item.Amount = int64(item.Quantity) * item.UnitPrice
		total += item.Amount
	}
	i.TotalAmount = total
	if i.Status == "" {
		i.Status = InboundPending
	}
	if !validInboundStatus(i.Status) {
		return NewValidationError("status", "入库单状态不合法")
	}
	return nil
}

func validInboundStatus(s string) bool {
	switch s {
	case InboundPending, InboundInspecting, InboundStocked, InboundRejected:
		return true
	}
	return false
}

// InboundFilter 入库单筛选条件。
type InboundFilter struct {
	Status          string
	PurchaseOrderID string
}

func (f InboundFilter) Match(i *InboundOrder) bool {
	if f.Status != "" && i.Status != f.Status {
		return false
	}
	if f.PurchaseOrderID != "" && i.PurchaseOrderID != f.PurchaseOrderID {
		return false
	}
	return true
}
