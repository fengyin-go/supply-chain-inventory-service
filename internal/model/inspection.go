package model

import (
	"strings"
	"time"
)

// 质检状态。
const (
	InspectionPending = "pending" // 待质检
	InspectionPassed  = "passed"  // 合格
	InspectionFailed  = "failed"  // 不合格
)

// inspectionTransitions 质检状态机。
var inspectionTransitions = map[string]map[string]bool{
	InspectionPending: {InspectionPassed: true, InspectionFailed: true},
}

// CanTransitionInspection 判断质检状态是否可流转。
func CanTransitionInspection(from, to string) bool {
	if m, ok := inspectionTransitions[from]; ok {
		return m[to]
	}
	return false
}

// Inspection 质检记录。
type Inspection struct {
	ID             string    `json:"id"`
	InboundOrderID string    `json:"inbound_order_id"`
	Inspector      string    `json:"inspector"`
	Result         string    `json:"result"` // pending/passed/failed
	DefectCount    int       `json:"defect_count"`
	Note           string    `json:"note"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (i *Inspection) Validate() error {
	i.InboundOrderID = strings.TrimSpace(i.InboundOrderID)
	i.Inspector = strings.TrimSpace(i.Inspector)
	i.Note = strings.TrimSpace(i.Note)
	if i.InboundOrderID == "" {
		return NewValidationError("inbound_order_id", "入库单 ID 不能为空")
	}
	if i.Inspector == "" {
		return NewValidationError("inspector", "质检员不能为空")
	}
	if i.DefectCount < 0 {
		return NewValidationError("defect_count", "缺陷数量不能为负数")
	}
	if i.Result == "" {
		i.Result = InspectionPending
	}
	if !validInspectionResult(i.Result) {
		return NewValidationError("result", "质检结果不合法")
	}
	return nil
}

func validInspectionResult(s string) bool {
	return s == InspectionPending || s == InspectionPassed || s == InspectionFailed
}

// InspectionFilter 质检记录筛选条件。
type InspectionFilter struct {
	Result         string
	InboundOrderID string
}

func (f InspectionFilter) Match(i *Inspection) bool {
	if f.Result != "" && i.Result != f.Result {
		return false
	}
	if f.InboundOrderID != "" && i.InboundOrderID != f.InboundOrderID {
		return false
	}
	return true
}
