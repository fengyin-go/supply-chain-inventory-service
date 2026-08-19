package model

import "testing"

func TestSupplierValidate(t *testing.T) {
	cases := []struct {
		name    string
		sup     Supplier
		wantErr bool
	}{
		{"正常", Supplier{Name: "华东电子", Contact: "王总", Phone: "13800000001", Address: "上海", Status: SupplierActive}, false},
		{"缺名称", Supplier{Name: " ", Phone: "13800000001"}, true},
		{"电话非法", Supplier{Name: "华东电子", Phone: "abc"}, true},
		{"状态非法", Supplier{Name: "华东电子", Phone: "13800000001", Status: "bogus"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.sup.Validate()
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestProductValidate(t *testing.T) {
	cases := []struct {
		name    string
		p       Product
		wantErr bool
	}{
		{"正常", Product{SKU: "SKU001", Name: "螺丝", Price: 100, LowStockThreshold: 10}, false},
		{"缺SKU", Product{SKU: " ", Name: "螺丝"}, true},
		{"缺名称", Product{SKU: "SKU001", Name: " "}, true},
		{"负价格", Product{SKU: "SKU001", Name: "螺丝", Price: -1}, true},
		{"负阈值", Product{SKU: "SKU001", Name: "螺丝", LowStockThreshold: -1}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.p.Validate()
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestPurchaseOrderValidate(t *testing.T) {
	validItems := []PurchaseItem{{ProductID: "p1", Quantity: 10, UnitPrice: 100}}
	cases := []struct {
		name    string
		po      PurchaseOrder
		wantErr bool
	}{
		{"正常", PurchaseOrder{SupplierID: "s1", Items: validItems}, false},
		{"缺供应商", PurchaseOrder{Items: validItems}, true},
		{"无明细", PurchaseOrder{SupplierID: "s1"}, true},
		{"数量为零", PurchaseOrder{SupplierID: "s1", Items: []PurchaseItem{{ProductID: "p1", Quantity: 0, UnitPrice: 100}}}, true},
		{"缺商品", PurchaseOrder{SupplierID: "s1", Items: []PurchaseItem{{Quantity: 1, UnitPrice: 100}}}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.po.Validate()
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestPurchaseOrderTotalAmount(t *testing.T) {
	po := PurchaseOrder{
		SupplierID: "s1",
		Items: []PurchaseItem{
			{ProductID: "p1", Quantity: 3, UnitPrice: 100},
			{ProductID: "p2", Quantity: 5, UnitPrice: 200},
		},
	}
	if err := po.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if po.TotalAmount != 1300 {
		t.Fatalf("total = %d, want 1300", po.TotalAmount)
	}
}

func TestStateMachines(t *testing.T) {
	// 采购单
	if !CanTransitionPO(POStatusDraft, POStatusConfirmed) || !CanTransitionPO(POStatusConfirmed, POStatusReceived) {
		t.Fatal("po transitions should be allowed")
	}
	if CanTransitionPO(POStatusReceived, POStatusConfirmed) {
		t.Fatal("received -> confirmed should be forbidden")
	}
	if CanTransitionPO(POStatusCancelled, POStatusConfirmed) {
		t.Fatal("cancelled -> confirmed should be forbidden")
	}

	// 入库单
	if !CanTransitionInbound(InboundPending, InboundInspecting) || !CanTransitionInbound(InboundInspecting, InboundStocked) {
		t.Fatal("inbound transitions should be allowed")
	}
	if CanTransitionInbound(InboundPending, InboundStocked) {
		t.Fatal("pending -> stocked should be forbidden")
	}
	if CanTransitionInbound(InboundStocked, InboundInspecting) {
		t.Fatal("stocked -> inspecting should be forbidden")
	}

	// 质检
	if !CanTransitionInspection(InspectionPending, InspectionPassed) || !CanTransitionInspection(InspectionPending, InspectionFailed) {
		t.Fatal("inspection transitions should be allowed")
	}
	if CanTransitionInspection(InspectionPassed, InspectionFailed) {
		t.Fatal("passed -> failed should be forbidden")
	}
}

func TestBatchAndMovementValidate(t *testing.T) {
	if err := (&InventoryBatch{BatchNo: "B1", ProductID: "p1", Quantity: 10, Remaining: 10}).Validate(); err != nil {
		t.Fatalf("batch valid: %v", err)
	}
	if err := (&InventoryBatch{BatchNo: "", ProductID: "p1", Quantity: 10}).Validate(); err == nil {
		t.Fatal("expected error for empty batch no")
	}
	if err := (&InventoryBatch{BatchNo: "B1", ProductID: "p1", Quantity: -1}).Validate(); err == nil {
		t.Fatal("expected error for negative quantity")
	}

	if err := (&StockMovement{ProductID: "p1", Type: MovementInbound, Delta: 5}).Validate(); err != nil {
		t.Fatalf("movement valid: %v", err)
	}
	if err := (&StockMovement{ProductID: "p1", Type: MovementInbound, Delta: 0}).Validate(); err == nil {
		t.Fatal("expected error for zero delta")
	}
	if err := (&StockMovement{ProductID: "p1", Type: "bogus", Delta: 5}).Validate(); err == nil {
		t.Fatal("expected error for bogus type")
	}
}
