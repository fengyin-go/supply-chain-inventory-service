package store

import "supplychain/internal/model"

func (s *MemoryStore) CreateReturnOrder(r *model.ReturnOrder) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.returns {
		if exist.ReturnNo == r.ReturnNo {
			return ErrConflict
		}
	}
	s.returns[r.ID] = r
	return nil
}

func (s *MemoryStore) GetReturnOrder(id string) (*model.ReturnOrder, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.returns[id]
	if !ok {
		return nil, ErrNotFound
	}
	return r, nil
}

func (s *MemoryStore) ListReturnOrders() []*model.ReturnOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.ReturnOrder, 0, len(s.returns))
	for _, r := range s.returns {
		list = append(list, r)
	}
	return list
}

func (s *MemoryStore) ListReturnOrdersByInbound(inboundID string) []*model.ReturnOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.ReturnOrder, 0)
	for _, r := range s.returns {
		if r.InboundOrderID == inboundID {
			list = append(list, r)
		}
	}
	return list
}

func (s *MemoryStore) UpdateReturnOrder(r *model.ReturnOrder) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.returns[r.ID]; !ok {
		return ErrNotFound
	}
	s.returns[r.ID] = r
	return nil
}

func (s *MemoryStore) CommitReturnCompletion(r *model.ReturnOrder) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.returns[r.ID]; !ok {
		return ErrNotFound
	}
	if r.Status != model.ReturnCompleted {
		return ErrConflict
	}
	s.returns[r.ID] = r
	return nil
}

func (s *MemoryStore) DeleteReturnOrder(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.returns[id]; !ok {
		return ErrNotFound
	}
	delete(s.returns, id)
	return nil
}
