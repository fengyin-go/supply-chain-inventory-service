package service

import (
	"sort"
	"time"

	"supplychain/internal/model"
	"supplychain/pkg/idgen"
)

// CreateSupplier 创建供应商。
func (s *Service) CreateSupplier(sup model.Supplier) (*model.Supplier, error) {
	if err := sup.Validate(); err != nil {
		return nil, err
	}
	sup.ID = idgen.Hex()
	now := time.Now()
	sup.CreatedAt = now
	sup.UpdatedAt = now
	if err := s.store.CreateSupplier(&sup); err != nil {
		return nil, err
	}
	return &sup, nil
}

func (s *Service) GetSupplier(id string) (*model.Supplier, error) {
	return s.store.GetSupplier(id)
}

func (s *Service) ListSuppliers(filter model.SupplierFilter, page, size int) ([]*model.Supplier, int, error) {
	all := s.store.ListSuppliers()
	matched := make([]*model.Supplier, 0, len(all))
	for _, sup := range all {
		if filter.Match(sup) {
			matched = append(matched, sup)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	return pageSlice(matched, page, size)
}

// UpdateSupplier 更新供应商可编辑字段。
func (s *Service) UpdateSupplier(id string, in model.Supplier) (*model.Supplier, error) {
	sup, err := s.store.GetSupplier(id)
	if err != nil {
		return nil, err
	}
	sup.Name = in.Name
	sup.Contact = in.Contact
	sup.Phone = in.Phone
	sup.Address = in.Address
	sup.Status = in.Status
	if err := sup.Validate(); err != nil {
		return nil, err
	}
	sup.UpdatedAt = time.Now()
	if err := s.store.UpdateSupplier(sup); err != nil {
		return nil, err
	}
	return sup, nil
}

func (s *Service) DeleteSupplier(id string) error {
	return s.store.DeleteSupplier(id)
}

// pageSlice 通用分页截取。
func pageSlice[T any](matched []T, page, size int) ([]T, int, error) {
	total := len(matched)
	start := (page - 1) * size
	if start >= total {
		return []T{}, total, nil
	}
	end := start + size
	if end > total {
		end = total
	}
	return matched[start:end], total, nil
}
