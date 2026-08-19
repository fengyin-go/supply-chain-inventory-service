package model

import "testing"

func TestSupplierFilterMatch(t *testing.T) {
	s := &Supplier{Name: "华东电子", Contact: "王总", Status: SupplierActive}
	cases := []struct {
		name   string
		filter SupplierFilter
		want   bool
	}{
		{"空筛选", SupplierFilter{}, true},
		{"状态匹配", SupplierFilter{Status: SupplierActive}, true},
		{"状态不匹配", SupplierFilter{Status: SupplierInactive}, false},
		{"关键词匹配名称", SupplierFilter{Keyword: "华东"}, true},
		{"关键词匹配联系人", SupplierFilter{Keyword: "王总"}, true},
		{"关键词不匹配", SupplierFilter{Keyword: "不存在"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.filter.Match(s); got != tc.want {
				t.Fatalf("Match = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestProductFilterMatch(t *testing.T) {
	p := &Product{SKU: "SKU001", Name: "不锈钢螺丝", Category: "紧固件"}
	cases := []struct {
		name   string
		filter ProductFilter
		want   bool
	}{
		{"分类匹配", ProductFilter{Category: "紧固件"}, true},
		{"分类不匹配", ProductFilter{Category: "五金"}, false},
		{"关键词匹配SKU", ProductFilter{Keyword: "sku"}, true},
		{"关键词匹配名称", ProductFilter{Keyword: "螺丝"}, true},
		{"关键词不匹配", ProductFilter{Keyword: "螺母"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.filter.Match(p); got != tc.want {
				t.Fatalf("Match = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestPurchaseOrderFilterMatch(t *testing.T) {
	po := &PurchaseOrder{SupplierID: "s1", Status: POStatusConfirmed}
	cases := []struct {
		name   string
		filter PurchaseOrderFilter
		want   bool
	}{
		{"状态匹配", PurchaseOrderFilter{Status: POStatusConfirmed}, true},
		{"状态不匹配", PurchaseOrderFilter{Status: POStatusDraft}, false},
		{"供应商匹配", PurchaseOrderFilter{SupplierID: "s1"}, true},
		{"供应商不匹配", PurchaseOrderFilter{SupplierID: "s2"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.filter.Match(po); got != tc.want {
				t.Fatalf("Match = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestInboundFilterMatch(t *testing.T) {
	in := &InboundOrder{PurchaseOrderID: "po1", Status: InboundPending}
	if !(InboundFilter{}).Match(in) {
		t.Fatal("empty filter should match")
	}
	if (InboundFilter{Status: InboundStocked}).Match(in) {
		t.Fatal("stocked filter should not match pending")
	}
	if !(InboundFilter{PurchaseOrderID: "po1"}).Match(in) {
		t.Fatal("po filter should match")
	}
}

func TestInspectionFilterMatch(t *testing.T) {
	ins := &Inspection{InboundOrderID: "in1", Result: InspectionPassed}
	if !(InspectionFilter{}).Match(ins) {
		t.Fatal("empty filter should match")
	}
	if !(InspectionFilter{Result: InspectionPassed}).Match(ins) {
		t.Fatal("passed filter should match")
	}
	if (InspectionFilter{Result: InspectionFailed}).Match(ins) {
		t.Fatal("failed filter should not match passed")
	}
}

func TestBatchAndMovementFilterMatch(t *testing.T) {
	b := &InventoryBatch{ProductID: "p1"}
	if !(BatchFilter{}).Match(b) {
		t.Fatal("empty batch filter should match")
	}
	if !(BatchFilter{ProductID: "p1"}).Match(b) {
		t.Fatal("product filter should match")
	}
	if (BatchFilter{ProductID: "p2"}).Match(b) {
		t.Fatal("wrong product filter should not match")
	}

	m := &StockMovement{ProductID: "p1", Type: MovementOutbound}
	if !(MovementFilter{}).Match(m) {
		t.Fatal("empty movement filter should match")
	}
	if !(MovementFilter{ProductID: "p1", Type: MovementOutbound}).Match(m) {
		t.Fatal("product+type filter should match")
	}
	if (MovementFilter{Type: MovementInbound}).Match(m) {
		t.Fatal("inbound filter should not match outbound")
	}
}

func TestReturnFilterMatch(t *testing.T) {
	r := &ReturnOrder{InboundOrderID: "in1", Status: ReturnPending}
	if !(ReturnFilter{}).Match(r) {
		t.Fatal("empty return filter should match")
	}
	if !(ReturnFilter{Status: ReturnPending, InboundOrderID: "in1"}).Match(r) {
		t.Fatal("status+inbound filter should match")
	}
	if (ReturnFilter{Status: ReturnCompleted}).Match(r) {
		t.Fatal("completed filter should not match pending")
	}
}
