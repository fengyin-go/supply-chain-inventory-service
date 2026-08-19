package store

import "supplychain/internal/model"

func (s *MemoryStore) CreateSupplier(sup *model.Supplier) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.suppliers {
		if exist.Name == sup.Name {
			return ErrConflict
		}
	}
	s.suppliers[sup.ID] = sup
	return nil
}

func (s *MemoryStore) GetSupplier(id string) (*model.Supplier, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sup, ok := s.suppliers[id]
	if !ok {
		return nil, ErrNotFound
	}
	return sup, nil
}

func (s *MemoryStore) ListSuppliers() []*model.Supplier {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Supplier, 0, len(s.suppliers))
	for _, sup := range s.suppliers {
		list = append(list, sup)
	}
	return list
}

func (s *MemoryStore) UpdateSupplier(sup *model.Supplier) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.suppliers[sup.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.suppliers {
		if exist.ID != sup.ID && exist.Name == sup.Name {
			return ErrConflict
		}
	}
	s.suppliers[sup.ID] = sup
	return nil
}

func (s *MemoryStore) DeleteSupplier(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.suppliers[id]; !ok {
		return ErrNotFound
	}
	delete(s.suppliers, id)
	return nil
}
