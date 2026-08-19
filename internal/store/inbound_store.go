package store

import "supplychain/internal/model"

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

func (s *MemoryStore) GetInboundOrderForStock(id string) (*model.InboundOrder, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	inbound, ok := s.inbounds[id]
	if !ok {
		return nil, ErrNotFound
	}
	return inbound, nil
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
