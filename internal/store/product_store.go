package store

import "supplychain/internal/model"

func (s *MemoryStore) CreateProduct(p *model.Product) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, exist := range s.products {
		if exist.SKU == p.SKU {
			return ErrConflict
		}
	}
	s.products[p.ID] = p
	return nil
}

func (s *MemoryStore) GetProduct(id string) (*model.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.products[id]
	if !ok {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *MemoryStore) GetProductBySKU(sku string) (*model.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.products {
		if p.SKU == sku {
			return p, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) ListProducts() []*model.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Product, 0, len(s.products))
	for _, p := range s.products {
		list = append(list, p)
	}
	return list
}

func (s *MemoryStore) UpdateProduct(p *model.Product) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.products[p.ID]; !ok {
		return ErrNotFound
	}
	for _, exist := range s.products {
		if exist.ID != p.ID && exist.SKU == p.SKU {
			return ErrConflict
		}
	}
	s.products[p.ID] = p
	return nil
}

func (s *MemoryStore) DeleteProduct(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.products[id]; !ok {
		return ErrNotFound
	}
	delete(s.products, id)
	return nil
}
