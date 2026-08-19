package service

import (
	"sort"
	"time"

	"supplychain/internal/model"
	"supplychain/pkg/idgen"
)

// CreateInboundOrder 从已确认的采购单创建入库单（待入库）。
func (s *Service) CreateInboundOrder(purchaseOrderID string) (*model.InboundOrder, error) {
	po, err := s.store.GetPurchaseOrder(purchaseOrderID)
	if err != nil {
		return nil, err
	}
	if po.Status != model.POStatusConfirmed {
		return nil, model.NewValidationError("purchase_order_id", "仅已确认的采购单可创建入库单")
	}
	// 同一采购单不可重复入库
	if existing := s.store.ListInboundOrdersByPurchaseOrder(purchaseOrderID); len(existing) > 0 {
		return nil, model.NewValidationError("purchase_order_id", "该采购单已存在入库单")
	}
	inbound := &model.InboundOrder{
		ID:              idgen.Hex(),
		InboundNo:       generateNo("IN"),
		PurchaseOrderID: po.ID,
		SupplierID:      po.SupplierID,
		Items:           model.CopyPurchaseItems(po.Items),
		TotalAmount:     po.TotalAmount,
		Status:          model.InboundPending,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := inbound.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.CreateInboundOrder(inbound); err != nil {
		return nil, err
	}
	return inbound, nil
}

func (s *Service) GetInboundOrder(id string) (*model.InboundOrder, error) {
	return s.store.GetInboundOrder(id)
}

func (s *Service) ListInboundOrders(filter model.InboundFilter, page, size int) ([]*model.InboundOrder, int, error) {
	all := s.store.ListInboundOrders()
	matched := make([]*model.InboundOrder, 0, len(all))
	for _, i := range all {
		if filter.Match(i) {
			matched = append(matched, i)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	return pageSlice(matched, page, size)
}

// StartInspection 开始质检：待入库 → 质检中，并创建待质检记录。
func (s *Service) StartInspection(inboundID, inspector string) (*model.Inspection, error) {
	inbound, err := s.store.GetInboundOrder(inboundID)
	if err != nil {
		return nil, err
	}
	if !model.CanTransitionInbound(inbound.Status, model.InboundInspecting) {
		return nil, model.NewValidationError("status", "当前状态不可开始质检")
	}
	inbound.Status = model.InboundInspecting
	inbound.UpdatedAt = time.Now()
	if err := s.store.UpdateInboundOrder(inbound); err != nil {
		return nil, err
	}
	ins := &model.Inspection{
		ID:             idgen.Hex(),
		InboundOrderID: inboundID,
		Inspector:      inspector,
		Result:         model.InspectionPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := ins.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.CreateInspection(ins); err != nil {
		return nil, err
	}
	return ins, nil
}

// RejectInbound 驳回入库单：待入库/质检中 → 已驳回。
func (s *Service) RejectInbound(id string) (*model.InboundOrder, error) {
	inbound, err := s.store.GetInboundOrder(id)
	if err != nil {
		return nil, err
	}
	if !model.CanTransitionInbound(inbound.Status, model.InboundRejected) {
		return nil, model.NewValidationError("status", "当前状态不可驳回")
	}
	inbound.Status = model.InboundRejected
	inbound.UpdatedAt = time.Now()
	if err := s.store.UpdateInboundOrder(inbound); err != nil {
		return nil, err
	}
	return inbound, nil
}

func (s *Service) DeleteInboundOrder(id string) error {
	return s.store.DeleteInboundOrder(id)
}
