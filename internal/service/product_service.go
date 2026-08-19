package service

import (
	"sort"
	"time"

	"supplychain/internal/model"
	"supplychain/pkg/idgen"
)

// CreateProduct 创建商品。
func (s *Service) CreateProduct(p model.Product) (*model.Product, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	p.ID = idgen.Hex()
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	if err := s.store.CreateProduct(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Service) GetProduct(id string) (*model.Product, error) {
	return s.store.GetProduct(id)
}

func (s *Service) ListProducts(filter model.ProductFilter, page, size int) ([]*model.Product, int, error) {
	all := s.store.ListProducts()
	matched := make([]*model.Product, 0, len(all))
	for _, p := range all {
		if filter.Match(p) {
			matched = append(matched, p)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	return pageSlice(matched, page, size)
}

// UpdateProduct 更新商品可编辑字段。
func (s *Service) UpdateProduct(id string, in model.Product) (*model.Product, error) {
	p, err := s.store.GetProduct(id)
	if err != nil {
		return nil, err
	}
	p.SKU = in.SKU
	p.Name = in.Name
	p.Category = in.Category
	p.Unit = in.Unit
	p.Price = in.Price
	p.LowStockThreshold = in.LowStockThreshold
	if err := p.Validate(); err != nil {
		return nil, err
	}
	p.UpdatedAt = time.Now()
	if err := s.store.UpdateProduct(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) DeleteProduct(id string) error {
	return s.store.DeleteProduct(id)
}
