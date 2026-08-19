package model

import "testing"

func TestReturnOrderValidate(t *testing.T) {
	cases := []struct {
		name    string
		ret     ReturnOrder
		wantErr bool
	}{
		{"正常", ReturnOrder{InboundOrderID: "in1", ProductID: "p1", Quantity: 2, Status: ReturnPending}, false},
		{"缺入库单", ReturnOrder{ProductID: "p1", Quantity: 2}, true},
		{"缺商品", ReturnOrder{InboundOrderID: "in1", Quantity: 2}, true},
		{"数量为零", ReturnOrder{InboundOrderID: "in1", ProductID: "p1", Quantity: 0}, true},
		{"状态非法", ReturnOrder{InboundOrderID: "in1", ProductID: "p1", Quantity: 2, Status: "bogus"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.ret.Validate()
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestInboundOrderValidate(t *testing.T) {
	validItems := []PurchaseItem{{ProductID: "p1", Quantity: 5, UnitPrice: 100}}
	cases := []struct {
		name    string
		in      InboundOrder
		wantErr bool
	}{
		{"正常", InboundOrder{PurchaseOrderID: "po1", Items: validItems}, false},
		{"缺采购单", InboundOrder{Items: validItems}, true},
		{"无明细", InboundOrder{PurchaseOrderID: "po1"}, true},
		{"数量为零", InboundOrder{PurchaseOrderID: "po1", Items: []PurchaseItem{{ProductID: "p1", Quantity: 0}}}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.in.Validate()
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestInspectionValidate(t *testing.T) {
	cases := []struct {
		name    string
		ins     Inspection
		wantErr bool
	}{
		{"正常", Inspection{InboundOrderID: "in1", Inspector: "小王", Result: InspectionPending}, false},
		{"缺入库单", Inspection{Inspector: "小王"}, true},
		{"缺质检员", Inspection{InboundOrderID: "in1"}, true},
		{"负缺陷", Inspection{InboundOrderID: "in1", Inspector: "小王", DefectCount: -1}, true},
		{"结果非法", Inspection{InboundOrderID: "in1", Inspector: "小王", Result: "bogus"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.ins.Validate()
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestStateMachineExhaustive 穷举所有状态机的合法/非法流转。
func TestStateMachineExhaustive(t *testing.T) {
	poStatuses := []string{POStatusDraft, POStatusConfirmed, POStatusReceived, POStatusCancelled}
	for _, from := range poStatuses {
		for _, to := range poStatuses {
			got := CanTransitionPO(from, to)
			want := false
			switch from {
			case POStatusDraft:
				want = to == POStatusConfirmed || to == POStatusCancelled
			case POStatusConfirmed:
				want = to == POStatusReceived
			}
			if got != want {
				t.Fatalf("PO %s -> %s = %v, want %v", from, to, got, want)
			}
		}
	}

	inboundStatuses := []string{InboundPending, InboundInspecting, InboundStocked, InboundRejected}
	for _, from := range inboundStatuses {
		for _, to := range inboundStatuses {
			got := CanTransitionInbound(from, to)
			want := false
			switch from {
			case InboundPending:
				want = to == InboundInspecting || to == InboundRejected
			case InboundInspecting:
				want = to == InboundStocked || to == InboundRejected
			}
			if got != want {
				t.Fatalf("Inbound %s -> %s = %v, want %v", from, to, got, want)
			}
		}
	}

	inspectionResults := []string{InspectionPending, InspectionPassed, InspectionFailed}
	for _, from := range inspectionResults {
		for _, to := range inspectionResults {
			got := CanTransitionInspection(from, to)
			want := from == InspectionPending && (to == InspectionPassed || to == InspectionFailed)
			if got != want {
				t.Fatalf("Inspection %s -> %s = %v, want %v", from, to, got, want)
			}
		}
	}

	returnStatuses := []string{ReturnPending, ReturnCompleted}
	for _, from := range returnStatuses {
		for _, to := range returnStatuses {
			got := CanTransitionReturn(from, to)
			want := from == ReturnPending && to == ReturnCompleted
			if got != want {
				t.Fatalf("Return %s -> %s = %v, want %v", from, to, got, want)
			}
		}
	}
}
