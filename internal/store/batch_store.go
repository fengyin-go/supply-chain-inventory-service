package store

import "supplychain/internal/model"

func (s *MemoryStore) CreateBatch(b *model.InventoryBatch) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.batches[b.ID] = b
	return nil
}

func (s *MemoryStore) GetBatch(id string) (*model.InventoryBatch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.batches[id]
	if !ok {
		return nil, ErrNotFound
	}
	return b, nil
}

func (s *MemoryStore) ListBatches() []*model.InventoryBatch {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.InventoryBatch, 0, len(s.batches))
	for _, b := range s.batches {
		list = append(list, b)
	}
	return list
}

func (s *MemoryStore) ListBatchesByProduct(productID string) []*model.InventoryBatch {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.InventoryBatch, 0)
	for _, b := range s.batches {
		if b.ProductID == productID {
			list = append(list, b)
		}
	}
	return list
}

func (s *MemoryStore) ListBatchesForOutbound(productID string) []*model.InventoryBatch {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.InventoryBatch, 0)
	for _, b := range s.batches {
		if b.ProductID == productID {
			cp := *b
			list = append(list, &cp)
		}
	}
	return list
}

func (s *MemoryStore) UpdateBatch(b *model.InventoryBatch) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.batches[b.ID]; !ok {
		return ErrNotFound
	}
	s.batches[b.ID] = b
	return nil
}

func (s *MemoryStore) DeleteBatch(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.batches[id]; !ok {
		return ErrNotFound
	}
	delete(s.batches, id)
	return nil
}
