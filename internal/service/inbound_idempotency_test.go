package service

import (
	"sync"
	"testing"

	"supplychain/internal/model"
	"supplychain/internal/store"
)

type stockInboundReadGate struct {
	store.Store
	mu      sync.Mutex
	enabled bool
	calls   int
	ready   chan struct{}
}

func (s *stockInboundReadGate) GetInboundOrder(id string) (*model.InboundOrder, error) {
	s.mu.Lock()
	if !s.enabled {
		s.mu.Unlock()
		return s.Store.GetInboundOrder(id)
	}
	s.calls++
	call := s.calls
	if call == 2 {
		close(s.ready)
	}
	s.mu.Unlock()
	if call <= 2 {
		<-s.ready
	}
	return s.Store.GetInboundOrder(id)
}

func TestConcurrentStockInboundIsIdempotent(t *testing.T) {
	base := store.NewMemoryStore()
	gated := &stockInboundReadGate{Store: base, ready: make(chan struct{})}
	svc := New(gated, testLogger(), testConfig())
	_, _, po := setupConfirmedPO(t, svc)
	inbound, _ := svc.CreateInboundOrder(po.ID)
	ins, _ := svc.StartInspection(inbound.ID, "qa")
	_, _ = svc.CompleteInspection(ins.ID, model.InspectionPassed, 0, "ok")
	gated.enabled = true

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.StockInbound(inbound.ID)
			errs <- err
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
		t.Fatalf("stock inbound successes = %d, want 1", successes)
	}
	if got := len(base.ListBatches()); got != 1 {
		t.Fatalf("batches = %d, want 1", got)
	}
	if got := len(base.ListMovements()); got != 1 {
		t.Fatalf("movements = %d, want 1", got)
	}
}
