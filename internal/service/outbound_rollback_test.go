package service

import (
	"errors"
	"testing"

	"supplychain/internal/model"
	"supplychain/internal/store"
)

type outboundMovementFailureStore struct {
	store.Store
}

func (s *outboundMovementFailureStore) CreateMovement(m *model.StockMovement) error {
	if m.Type == model.MovementOutbound {
		return errors.New("movement writer unavailable")
	}
	return s.Store.CreateMovement(m)
}

func TestOutboundMovementFailureRestoresBatch(t *testing.T) {
	base := store.NewMemoryStore()
	svc := New(&outboundMovementFailureStore{Store: base}, testLogger(), testConfig())
	sup := seedSupplier(t, svc, "Dock", "13800000011")
	product := seedProduct(t, svc, "CAP-A", "capacitor", 100, 1)
	po := seedPurchaseOrder(t, svc, sup.ID, []model.PurchaseItem{{ProductID: product.ID, Quantity: 10, UnitPrice: 100}})
	_, _ = svc.ConfirmPurchaseOrder(po.ID)
	inbound, _ := svc.CreateInboundOrder(po.ID)
	ins, _ := svc.StartInspection(inbound.ID, "qa")
	_, _ = svc.CompleteInspection(ins.ID, model.InspectionPassed, 0, "ok")
	_, _ = svc.StockInbound(inbound.ID)

	if err := svc.OutboundStock(product.ID, 4, "order"); err == nil {
		t.Fatal("expected movement failure")
	}
	stock, _ := svc.GetStock(product.ID)
	if stock != 10 {
		t.Fatalf("stock after failed outbound = %d, want 10", stock)
	}
	if got := len(base.ListMovementsByProduct(product.ID)); got != 1 {
		t.Fatalf("movement count = %d, want inbound movement only", got)
	}
}
