package model

// 本文件为各实体提供深拷贝（Clone）。导出快照必须返回独立副本，
// 避免调用方修改快照时连带改动内存中的原始记录（隔离导出与系统数据）。

// Clone 返回 Supplier 的深拷贝。
func (s *Supplier) Clone() *Supplier {
	if s == nil {
		return nil
	}
	cp := *s
	return &cp
}

// Clone 返回 Product 的深拷贝。
func (p *Product) Clone() *Product {
	if p == nil {
		return nil
	}
	cp := *p
	return &cp
}

// Clone 返回 PurchaseOrder 的深拷贝，包含明细行。
func (p *PurchaseOrder) Clone() *PurchaseOrder {
	if p == nil {
		return nil
	}
	cp := *p
	if p.Items != nil {
		cp.Items = make([]PurchaseItem, len(p.Items))
		copy(cp.Items, p.Items)
	}
	return &cp
}

// Clone 返回 InboundOrder 的深拷贝，包含明细行。
func (i *InboundOrder) Clone() *InboundOrder {
	if i == nil {
		return nil
	}
	cp := *i
	if i.Items != nil {
		cp.Items = make([]PurchaseItem, len(i.Items))
		copy(cp.Items, i.Items)
	}
	return &cp
}

// Clone 返回 Inspection 的深拷贝。
func (i *Inspection) Clone() *Inspection {
	if i == nil {
		return nil
	}
	cp := *i
	return &cp
}

// Clone 返回 InventoryBatch 的深拷贝。
func (b *InventoryBatch) Clone() *InventoryBatch {
	if b == nil {
		return nil
	}
	cp := *b
	return &cp
}

// Clone 返回 StockMovement 的深拷贝。
func (m *StockMovement) Clone() *StockMovement {
	if m == nil {
		return nil
	}
	cp := *m
	return &cp
}

// CloneSlice 返回指针切片的深拷贝，对每个元素调用其 Clone 方法。
// nil 入参返回 nil，空切片返回空切片，以保留 JSON 序列化时的 [] 语义。
func CloneSlice[T any, PT interface {
	*T
	Clone() PT
}](in []PT) []PT {
	if in == nil {
		return nil
	}
	out := make([]PT, len(in))
	for i, v := range in {
		out[i] = v.Clone()
	}
	return out
}
