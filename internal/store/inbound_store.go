package store

import (
	"time"

	"supplychain/internal/model"
)

func (s *MemoryStore) CreateInboundOrder(i *model.InboundOrder) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.inbounds {
		if exist.InboundNo == i.InboundNo {
			return ErrConflict
		}
	}
	s.inbounds[i.ID] = i
	return nil
}

func (s *MemoryStore) GetInboundOrder(id string) (*model.InboundOrder, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	i, ok := s.inbounds[id]
	if !ok {
		return nil, ErrNotFound
	}
	return i, nil
}

func (s *MemoryStore) ListInboundOrders() []*model.InboundOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.InboundOrder, 0, len(s.inbounds))
	for _, i := range s.inbounds {
		list = append(list, i)
	}
	return list
}

func (s *MemoryStore) ListInboundOrdersByPurchaseOrder(poID string) []*model.InboundOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.InboundOrder, 0)
	for _, i := range s.inbounds {
		if i.PurchaseOrderID == poID {
			list = append(list, i)
		}
	}
	return list
}

func (s *MemoryStore) UpdateInboundOrder(i *model.InboundOrder) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.inbounds[i.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.inbounds {
		if exist.ID != i.ID && exist.InboundNo == i.InboundNo {
			return ErrConflict
		}
	}
	s.inbounds[i.ID] = i
	return nil
}

func (s *MemoryStore) DeleteInboundOrder(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.inbounds[id]; !ok {
		return ErrNotFound
	}
	delete(s.inbounds, id)
	return nil
}

// TransitionInboundStatus 原子地比较并流转入库单状态。
// 仅当入库单当前状态等于 from 时，才更新为 to 并返回更新后的入库单；
// 若状态已被并发请求改掉则返回 ErrConflict，调用方据此识别重复/并发记账，
// 避免在 Service 层因「读取状态 → 判定 → 写入」之间出现竞态而重复入库。
func (s *MemoryStore) TransitionInboundStatus(id string, from, to string) (*model.InboundOrder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	i, ok := s.inbounds[id]
	if !ok {
		return nil, ErrNotFound
	}
	if i.Status != from {
		return nil, ErrConflict
	}
	i.Status = to
	i.UpdatedAt = time.Now()
	return i, nil
}
