package service

import (
	"sync"
	"testing"

	"supplychain/internal/model"
	"supplychain/internal/store"
)

type outboundReadGate struct {
	store.Store
	mu    sync.Mutex
	calls int
	ready chan struct{}
}

func (s *outboundReadGate) ListBatchesByProduct(productID string) []*model.InventoryBatch {
	s.mu.Lock()
	s.calls++
	call := s.calls
	if call == 2 {
		close(s.ready)
	}
	s.mu.Unlock()
	if call <= 2 {
		<-s.ready
	}
	return s.Store.ListBatchesByProduct(productID)
}

func TestConcurrentOutboundCannotOversellBatch(t *testing.T) {
	base := store.NewMemoryStore()
	svc := New(&outboundReadGate{Store: base, ready: make(chan struct{})}, testLogger(), testConfig())
	sup := seedSupplier(t, svc, "Dock", "13800000010")
	product := seedProduct(t, svc, "CHIP-A", "chip", 100, 1)
	po := seedPurchaseOrder(t, svc, sup.ID, []model.PurchaseItem{{ProductID: product.ID, Quantity: 10, UnitPrice: 100}})
	_, _ = svc.ConfirmPurchaseOrder(po.ID)
	inbound, _ := svc.CreateInboundOrder(po.ID)
	ins, _ := svc.StartInspection(inbound.ID, "qa")
	_, _ = svc.CompleteInspection(ins.ID, model.InspectionPassed, 0, "ok")
	_, _ = svc.StockInbound(inbound.ID)

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- svc.OutboundStock(product.ID, 8, "order")
		}()
	}
	wg.Wait()
	close(errs)
	successes := 0
	for err := range errs {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("outbound successes = %d, want 1", successes)
	}
	batches := base.ListBatchesByProduct(product.ID)
	stock := 0
	for _, batch := range batches {
		stock += batch.Remaining
	}
	if stock != 2 {
		t.Fatalf("stock after concurrent outbound = %d, want 2", stock)
	}
}
