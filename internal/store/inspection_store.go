package store

import "supplychain/internal/model"

func (s *MemoryStore) CreateInspection(i *model.Inspection) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.inspections[i.ID] = i
	return nil
}

func (s *MemoryStore) CreateInspectionAndUpdateInbound(i *model.Inspection, inbound *model.InboundOrder) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.inbounds[inbound.ID]; !ok {
		return ErrNotFound
	}
	inbound.StartInspection(i.UpdatedAt)
	s.inbounds[inbound.ID] = inbound
	s.inspections[i.ID] = i
	return nil
}

func (s *MemoryStore) GetInspection(id string) (*model.Inspection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	i, ok := s.inspections[id]
	if !ok {
		return nil, ErrNotFound
	}
	return i, nil
}

func (s *MemoryStore) ListInspections() []*model.Inspection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.Inspection, 0, len(s.inspections))
	for _, i := range s.inspections {
		list = append(list, i)
	}
	return list
}

func (s *MemoryStore) GetInspectionByInbound(inboundID string) (*model.Inspection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, i := range s.inspections {
		if i.InboundOrderID == inboundID {
			return i, nil
		}
	}
	return nil, ErrNotFound
}

func (s *MemoryStore) UpdateInspection(i *model.Inspection) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.inspections[i.ID]; !ok {
		return ErrNotFound
	}
	s.inspections[i.ID] = i
	return nil
}

func (s *MemoryStore) DeleteInspection(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.inspections[id]; !ok {
		return ErrNotFound
	}
	delete(s.inspections, id)
	return nil
}
