package store

import "supplychain/internal/model"

func (s *MemoryStore) CreateMovement(m *model.StockMovement) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.movements[m.ID] = m
	return nil
}

func (s *MemoryStore) GetMovement(id string) (*model.StockMovement, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.movements[id]
	if !ok {
		return nil, ErrNotFound
	}
	return m, nil
}

func (s *MemoryStore) ListMovements() []*model.StockMovement {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.StockMovement, 0, len(s.movements))
	for _, m := range s.movements {
		list = append(list, m)
	}
	return list
}

func (s *MemoryStore) ListMovementsByProduct(productID string) []*model.StockMovement {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.StockMovement, 0)
	for _, m := range s.movements {
		if m.ProductID == productID {
			list = append(list, m)
		}
	}
	return list
}

func (s *MemoryStore) DeleteMovement(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.movements[id]; !ok {
		return ErrNotFound
	}
	delete(s.movements, id)
	return nil
}

func (s *MemoryStore) RollbackMovement(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.movements, id)
	return nil
}
