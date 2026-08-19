package store

import "supplychain/internal/model"

func (s *MemoryStore) CreatePurchaseOrder(p *model.PurchaseOrder) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.pos {
		if exist.OrderNo == p.OrderNo {
			return ErrConflict
		}
	}
	s.pos[p.ID] = p
	return nil
}

func (s *MemoryStore) GetPurchaseOrder(id string) (*model.PurchaseOrder, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.pos[id]
	if !ok {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *MemoryStore) ListPurchaseOrders() []*model.PurchaseOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.PurchaseOrder, 0, len(s.pos))
	for _, p := range s.pos {
		list = append(list, p)
	}
	return list
}

func (s *MemoryStore) UpdatePurchaseOrder(p *model.PurchaseOrder) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.pos[p.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.pos {
		if exist.ID != p.ID && exist.OrderNo == p.OrderNo {
			return ErrConflict
		}
	}
	s.pos[p.ID] = p
	return nil
}

func (s *MemoryStore) DeletePurchaseOrder(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.pos[id]; !ok {
		return ErrNotFound
	}
	delete(s.pos, id)
	return nil
}
