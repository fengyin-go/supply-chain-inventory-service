package model

import (
	"strings"
	"time"
)

// Product 商品/物料。价格单位「分」。
type Product struct {
	ID                string    `json:"id"`
	SKU               string    `json:"sku"`
	Name              string    `json:"name"`
	Category          string    `json:"category"`
	Unit              string    `json:"unit"`
	Price             int64     `json:"price"`               // 单位分
	LowStockThreshold int       `json:"low_stock_threshold"` // 低库存阈值
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (p *Product) Snapshot() *Product {
	if p == nil {
		return nil
	}
	cp := *p
	return &cp
}

func (p *Product) Validate() error {
	p.SKU = strings.TrimSpace(p.SKU)
	p.Name = strings.TrimSpace(p.Name)
	p.Category = strings.TrimSpace(p.Category)
	p.Unit = strings.TrimSpace(p.Unit)
	if p.SKU == "" {
		return NewValidationError("sku", "SKU 不能为空")
	}
	if p.Name == "" {
		return NewValidationError("name", "商品名称不能为空")
	}
	if p.Price < 0 {
		return NewValidationError("price", "价格不能为负数")
	}
	if p.LowStockThreshold < 0 {
		return NewValidationError("low_stock_threshold", "低库存阈值不能为负数")
	}
	if p.Unit == "" {
		p.Unit = "件"
	}
	return nil
}

// ProductFilter 商品筛选条件。
type ProductFilter struct {
	Category string
	Keyword  string
}

func (f ProductFilter) Match(p *Product) bool {
	if f.Category != "" && p.Category != f.Category {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(p.Name), k) &&
			!strings.Contains(strings.ToLower(p.SKU), k) {
			return false
		}
	}
	return true
}
