package service

import (
	"sync"
	"testing"

	"supplychain/internal/config"
	"supplychain/internal/model"
	"supplychain/internal/store"
	"supplychain/pkg/logger"
)

type inboundCheckGate struct {
	store.Store
	mu    sync.Mutex
	calls int
	ready chan struct{}
}

func (g *inboundCheckGate) ListInboundOrdersByPurchaseOrder(poID string) []*model.InboundOrder {
	g.mu.Lock()
	g.calls++
	call := g.calls
	if call == 2 {
		close(g.ready)
	}
	g.mu.Unlock()
	if call <= 2 {
		<-g.ready
	}
	return g.Store.ListInboundOrdersByPurchaseOrder(poID)
}

func TestConcurrentInboundCreationKeepsOneOrder(t *testing.T) {
	base := store.NewMemoryStore()
	gated := &inboundCheckGate{Store: base, ready: make(chan struct{})}
	svc := New(gated, logger.NewLevel(logger.LevelError), &config.Config{MaxPageSize: 100})
	_, _, po := setupConfirmedPO(t, svc)

	var wg sync.WaitGroup
	results := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.CreateInboundOrder(po.ID)
			results <- err
		}()
	}
	wg.Wait()
	close(results)

	successes := 0
	for err := range results {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("concurrent creation successes = %d, want 1", successes)
	}
	inbounds := base.ListInboundOrdersByPurchaseOrder(po.ID)
	if len(inbounds) != 1 {
		t.Fatalf("stored inbounds = %d, want 1", len(inbounds))
	}
}
