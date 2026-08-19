package service

import (
	"errors"
	"testing"

	"supplychain/internal/config"
	"supplychain/internal/model"
	"supplychain/internal/store"
	"supplychain/pkg/logger"
)

type inspectionCreateFailureStore struct {
	store.Store
}

func (s *inspectionCreateFailureStore) CreateInspection(i *model.Inspection) error {
	return errors.New("inspection writer unavailable")
}

func (s *inspectionCreateFailureStore) CreateInspectionAndUpdateInbound(i *model.Inspection, in *model.InboundOrder) error {
	return errors.New("inspection writer unavailable")
}

func TestStartInspectionDoesNotLeaveInboundHalfUpdated(t *testing.T) {
	base := store.NewMemoryStore()
	svc := New(&inspectionCreateFailureStore{Store: base}, logger.NewLevel(logger.LevelError), &config.Config{MaxPageSize: 100})
	_, _, po := setupConfirmedPO(t, svc)
	inbound, err := svc.CreateInboundOrder(po.ID)
	if err != nil {
		t.Fatalf("create inbound: %v", err)
	}

	if _, err := svc.StartInspection(inbound.ID, "qa"); err == nil {
		t.Fatal("expected inspection creation failure")
	}
	current, err := svc.GetInboundOrder(inbound.ID)
	if err != nil {
		t.Fatalf("get inbound: %v", err)
	}
	if current.Status != model.InboundPending {
		t.Fatalf("inbound status = %q, want pending after failed start", current.Status)
	}
	if got := len(base.ListInspections()); got != 0 {
		t.Fatalf("stored inspections = %d, want 0", got)
	}
}
