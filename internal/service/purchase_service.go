package service

import (
	"fmt"
	"sort"
	"time"

	"supplychain/internal/model"
	"supplychain/pkg/idgen"
)

// CreatePurchaseOrder 创建采购单（草稿）。
func (s *Service) CreatePurchaseOrder(in model.PurchaseOrder) (*model.PurchaseOrder, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}
	// 跨实体校验：供应商存在
	if _, err := s.store.GetSupplier(in.SupplierID); err != nil {
		return nil, err
	}
	// 跨实体校验：每个明细商品存在
	if err := s.validateItems(in.Items); err != nil {
		return nil, err
	}
	if in.OrderNo == "" {
		in.OrderNo = generateNo("PO")
	}
	in.ID = idgen.Hex()
	in.Status = model.POStatusDraft
	now := time.Now()
	in.CreatedAt = now
	in.UpdatedAt = now
	if err := s.store.CreatePurchaseOrder(&in); err != nil {
		return nil, err
	}
	return &in, nil
}

// validateItems 校验明细中的商品均存在。
func (s *Service) validateItems(items []model.PurchaseItem) error {
	for _, item := range items {
		if _, err := s.store.GetProduct(item.ProductID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) GetPurchaseOrder(id string) (*model.PurchaseOrder, error) {
	return s.store.GetPurchaseOrder(id)
}

func (s *Service) ListPurchaseOrders(filter model.PurchaseOrderFilter, page, size int) ([]*model.PurchaseOrder, int, error) {
	all := s.store.ListPurchaseOrders()
	matched := make([]*model.PurchaseOrder, 0, len(all))
	for _, p := range all {
		if filter.Match(p) {
			matched = append(matched, p)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	return pageSlice(matched, page, size)
}

// ConfirmPurchaseOrder 确认采购单：草稿 → 已确认。
func (s *Service) ConfirmPurchaseOrder(id string) (*model.PurchaseOrder, error) {
	p, err := s.store.GetPurchaseOrder(id)
	if err != nil {
		return nil, err
	}
	if !model.CanTransitionPO(p.Status, model.POStatusConfirmed) {
		return nil, model.NewValidationError("status", "当前状态不可确认")
	}
	p.Status = model.POStatusConfirmed
	p.UpdatedAt = time.Now()
	if err := s.store.UpdatePurchaseOrder(p); err != nil {
		return nil, err
	}
	return p, nil
}

// ReceivePurchaseOrder 收货完成：已确认 → 已收货（由入库完成时触发）。
func (s *Service) receivePurchaseOrder(id string) error {
	p, err := s.store.GetPurchaseOrder(id)
	if err != nil {
		return err
	}
	if !model.CanTransitionPO(p.Status, model.POStatusReceived) {
		return model.NewValidationError("status", "当前状态不可标记收货")
	}
	p.Status = model.POStatusReceived
	p.UpdatedAt = time.Now()
	return s.store.UpdatePurchaseOrder(p)
}

// CancelPurchaseOrder 取消采购单：草稿 → 已取消。
func (s *Service) CancelPurchaseOrder(id string) (*model.PurchaseOrder, error) {
	p, err := s.store.GetPurchaseOrder(id)
	if err != nil {
		return nil, err
	}
	if !model.CanTransitionPO(p.Status, model.POStatusCancelled) {
		return nil, model.NewValidationError("status", "当前状态不可取消")
	}
	p.Status = model.POStatusCancelled
	p.UpdatedAt = time.Now()
	if err := s.store.UpdatePurchaseOrder(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) DeletePurchaseOrder(id string) error {
	return s.store.DeletePurchaseOrder(id)
}

// generateNo 生成单号，形如 PO-20260816-abc123。
func generateNo(prefix string) string {
	return fmt.Sprintf("%s-%s-%s", prefix, time.Now().Format("20060102"), idgen.Short())
}
