package model

import (
	"regexp"
	"strings"
	"time"
)

// 供应商状态。
const (
	SupplierActive   = "active"
	SupplierInactive = "inactive"
)

// Supplier 供应商。
type Supplier struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Contact   string    `json:"contact"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *Supplier) Snapshot() *Supplier {
	if s == nil {
		return nil
	}
	cp := *s
	return &cp
}

var phoneRe = regexp.MustCompile(`^\d{6,15}$`)

func (s *Supplier) Validate() error {
	s.Name = strings.TrimSpace(s.Name)
	s.Contact = strings.TrimSpace(s.Contact)
	s.Phone = strings.TrimSpace(s.Phone)
	s.Address = strings.TrimSpace(s.Address)
	if s.Name == "" {
		return NewValidationError("name", "供应商名称不能为空")
	}
	if !phoneRe.MatchString(s.Phone) {
		return NewValidationError("phone", "联系电话格式不合法")
	}
	if s.Status == "" {
		s.Status = SupplierActive
	}
	if s.Status != SupplierActive && s.Status != SupplierInactive {
		return NewValidationError("status", "供应商状态不合法")
	}
	return nil
}

// SupplierFilter 供应商筛选条件。
type SupplierFilter struct {
	Status  string
	Keyword string
}

func (f SupplierFilter) Match(s *Supplier) bool {
	if f.Status != "" && s.Status != f.Status {
		return false
	}
	if f.Keyword != "" {
		k := strings.ToLower(strings.TrimSpace(f.Keyword))
		if k != "" && !strings.Contains(strings.ToLower(s.Name), k) &&
			!strings.Contains(strings.ToLower(s.Contact), k) {
			return false
		}
	}
	return true
}
