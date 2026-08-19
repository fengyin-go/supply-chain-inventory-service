package model

import (
	"strings"
	"time"
)

// 采购单状态。
const (
	POStatusDraft     = "draft"     // 草稿
	POStatusConfirmed = "confirmed" // 已确认
	POStatusReceived  = "received"  // 已收货
	POStatusCancelled = "cancelled" // 已取消
)

// poTransitions 采购单状态机。
var poTransitions = map[string]map[string]bool{
	POStatusDraft:     {POStatusConfirmed: true, POStatusCancelled: true},
	POStatusConfirmed: {POStatusReceived: true},
}

// CanTransitionPO 判断采购单状态是否可流转。
func CanTransitionPO(from, to string) bool {
	if m, ok := poTransitions[from]; ok {
		return m[to]
	}
	return false
}

// PurchaseItem 采购明细行。
type PurchaseItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	UnitPrice int64  `json:"unit_price"` // 单位分
	Amount    int64  `json:"amount"`     // 行金额 = 数量 × 单价
}

// PurchaseOrder 采购单。金额单位「分」。
type PurchaseOrder struct {
	ID          string         `json:"id"`
	OrderNo     string         `json:"order_no"`
	SupplierID  string         `json:"supplier_id"`
	Items       []PurchaseItem `json:"items"`
	TotalAmount int64          `json:"total_amount"`
	Status      string         `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func (p *PurchaseOrder) Snapshot() *PurchaseOrder {
	if p == nil {
		return nil
	}
	cp := *p
	cp.Items = append([]PurchaseItem(nil), p.Items...)
	return &cp
}

func (p *PurchaseOrder) Validate() error {
	p.OrderNo = strings.TrimSpace(p.OrderNo)
	p.SupplierID = strings.TrimSpace(p.SupplierID)
	if p.SupplierID == "" {
		return NewValidationError("supplier_id", "供应商 ID 不能为空")
	}
	if len(p.Items) == 0 {
		return NewValidationError("items", "采购明细不能为空")
	}
	var total int64
	for i := range p.Items {
		item := &p.Items[i]
		item.ProductID = strings.TrimSpace(item.ProductID)
		if item.ProductID == "" {
			return NewValidationError("items", "明细中的商品 ID 不能为空")
		}
		if item.Quantity <= 0 {
			return NewValidationError("items", "明细中的数量必须大于 0")
		}
		if item.UnitPrice < 0 {
			return NewValidationError("items", "明细中的单价不能为负数")
		}
		item.Amount = int64(item.Quantity) * item.UnitPrice
		total += item.Amount
	}
	p.TotalAmount = total
	if p.Status == "" {
		p.Status = POStatusDraft
	}
	if !validPOStatus(p.Status) {
		return NewValidationError("status", "采购单状态不合法")
	}
	return nil
}

func validPOStatus(s string) bool {
	switch s {
	case POStatusDraft, POStatusConfirmed, POStatusReceived, POStatusCancelled:
		return true
	}
	return false
}

// PurchaseOrderFilter 采购单筛选条件。
type PurchaseOrderFilter struct {
	Status     string
	SupplierID string
}

func (f PurchaseOrderFilter) Match(p *PurchaseOrder) bool {
	if f.Status != "" && p.Status != f.Status {
		return false
	}
	if f.SupplierID != "" && p.SupplierID != f.SupplierID {
		return false
	}
	return true
}
