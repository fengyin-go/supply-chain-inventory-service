package service

import (
	"sort"
	"time"

	"supplychain/internal/model"
	"supplychain/pkg/idgen"
)

// CreateReturnOrder 创建采购退货单（待处理）。
// 要求入库单已入库，且退货商品属于该入库单明细。
func (s *Service) CreateReturnOrder(inboundOrderID, productID string, quantity int, reason string) (*model.ReturnOrder, error) {
	inbound, err := s.store.GetInboundOrder(inboundOrderID)
	if err != nil {
		return nil, err
	}
	if inbound.Status != model.InboundStocked {
		return nil, model.NewValidationError("inbound_order_id", "仅已入库的入库单可退货")
	}
	found := false
	maxQty := 0
	for _, item := range inbound.Items {
		if item.ProductID == productID {
			found = true
			maxQty = item.Quantity
			break
		}
	}
	if !found {
		return nil, model.NewValidationError("product_id", "退货商品不属于该入库单")
	}
	if quantity > maxQty {
		return nil, model.NewValidationError("quantity", "退货数量超过入库数量")
	}
	ret := &model.ReturnOrder{
		ID:             idgen.Hex(),
		ReturnNo:       generateNo("RT"),
		InboundOrderID: inboundOrderID,
		ProductID:      productID,
		Quantity:       quantity,
		Reason:         reason,
		Status:         model.ReturnPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := ret.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.CreateReturnOrder(ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *Service) GetReturnOrder(id string) (*model.ReturnOrder, error) {
	return s.store.GetReturnOrder(id)
}

func (s *Service) ListReturnOrders(filter model.ReturnFilter, page, size int) ([]*model.ReturnOrder, int, error) {
	all := s.store.ListReturnOrders()
	matched := make([]*model.ReturnOrder, 0, len(all))
	for _, r := range all {
		if filter.Match(r) {
			matched = append(matched, r)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	return pageSlice(matched, page, size)
}

// CompleteReturnOrder 完成退货：待处理 → 已完成，并扣减库存。
func (s *Service) CompleteReturnOrder(id string) (*model.ReturnOrder, error) {
	ret, err := s.store.GetReturnOrder(id)
	if err != nil {
		return nil, err
	}
	if !model.CanTransitionReturn(ret.Status, model.ReturnCompleted) {
		return nil, model.NewValidationError("status", "当前状态不可完成退货")
	}
	// 扣减库存（先进先出）
	if err := s.OutboundStock(ret.ProductID, ret.Quantity, "退货:"+ret.ReturnNo); err != nil {
		return nil, err
	}
	ret.Status = model.ReturnCompleted
	ret.UpdatedAt = time.Now()
	if err := s.store.UpdateReturnOrder(ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func (s *Service) DeleteReturnOrder(id string) error {
	return s.store.DeleteReturnOrder(id)
}

// ReturnStats 退货统计。
type ReturnStats struct {
	Total          int `json:"total"`
	Pending        int `json:"pending"`
	Completed      int `json:"completed"`
	TotalQuantity  int `json:"total_quantity"`
}

func (s *Service) ReturnStats() (*ReturnStats, error) {
	returns := s.store.ListReturnOrders()
	stats := &ReturnStats{Total: len(returns)}
	for _, r := range returns {
		stats.TotalQuantity += r.Quantity
		switch r.Status {
		case model.ReturnPending:
			stats.Pending++
		case model.ReturnCompleted:
			stats.Completed++
		}
	}
	return stats, nil
}
