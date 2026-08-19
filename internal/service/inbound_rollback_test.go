package service

import (
	"errors"
	"testing"

	"supplychain/internal/model"
	"supplychain/internal/store"
)

type movementFailureStore struct {
	store.Store
	creates int
}

func (s *movementFailureStore) CreateMovement(m *model.StockMovement) error {
	s.creates++
	if s.creates == 2 {
		return errors.New("movement writer unavailable")
	}
	return s.Store.CreateMovement(m)
}

func TestPartialInboundFailureLeavesNoInventoryArtifacts(t *testing.T) {
	base := store.NewMemoryStore()
	svc := New(&movementFailureStore{Store: base}, testLogger(), testConfig())
	sup := seedSupplier(t, svc, "Dock", "13800000009")
	p1 := seedProduct(t, svc, "ROLL-A", "roll A", 100, 1)
	p2 := seedProduct(t, svc, "ROLL-B", "roll B", 200, 1)
	po := seedPurchaseOrder(t, svc, sup.ID, []model.PurchaseItem{
		{ProductID: p1.ID, Quantity: 3, UnitPrice: 100},
		{ProductID: p2.ID, Quantity: 4, UnitPrice: 200},
	})
	_, _ = svc.ConfirmPurchaseOrder(po.ID)
	inbound, _ := svc.CreateInboundOrder(po.ID)
	ins, _ := svc.StartInspection(inbound.ID, "qa")
	_, _ = svc.CompleteInspection(ins.ID, model.InspectionPassed, 0, "ok")

	if _, err := svc.StockInbound(inbound.ID); err == nil {
		t.Fatal("expected movement failure")
	}
	if got := len(base.ListBatches()); got != 0 {
		t.Fatalf("batches after failed inbound = %d, want 0", got)
	}
	if got := len(base.ListMovements()); got != 0 {
		t.Fatalf("movements after failed inbound = %d, want 0", got)
	}
	current, _ := svc.GetInboundOrder(inbound.ID)
	if current.Status != model.InboundInspecting {
		t.Fatalf("inbound status = %q, want inspecting", current.Status)
	}
}
