package service

import (
	"errors"
	"testing"

	"supplychain/internal/model"
	"supplychain/internal/store"
)

type returnUpdateFailureStore struct {
	store.Store
}

func (s *returnUpdateFailureStore) UpdateReturnOrder(r *model.ReturnOrder) error {
	return errors.New("return writer unavailable")
}

func TestFailedReturnCommitDoesNotConsumeStock(t *testing.T) {
	svc := New(&returnUpdateFailureStore{Store: store.NewMemoryStore()}, testLogger(), testConfig())
	_, product, po := setupConfirmedPO(t, svc)
	inbound, _ := svc.CreateInboundOrder(po.ID)
	ins, _ := svc.StartInspection(inbound.ID, "qa")
	_, _ = svc.CompleteInspection(ins.ID, model.InspectionPassed, 0, "ok")
	if _, err := svc.StockInbound(inbound.ID); err != nil {
		t.Fatalf("stock inbound: %v", err)
	}
	ret, err := svc.CreateReturnOrder(inbound.ID, product.ID, 4, "damaged")
	if err != nil {
		t.Fatalf("create return: %v", err)
	}

	if _, err := svc.CompleteReturnOrder(ret.ID); err == nil {
		t.Fatal("expected return update failure")
	}
	stock, err := svc.GetStock(product.ID)
	if err != nil {
		t.Fatalf("get stock: %v", err)
	}
	if stock != 10 {
		t.Fatalf("stock after failed return = %d, want 10", stock)
	}
	current, _ := svc.GetReturnOrder(ret.ID)
	if current.Status != model.ReturnPending {
		t.Fatalf("return status = %q, want pending", current.Status)
	}
}
