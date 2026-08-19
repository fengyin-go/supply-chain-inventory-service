package service

import (
	"testing"

	"supplychain/internal/model"
)

func TestReturnOrdersReserveInboundQuantity(t *testing.T) {
	svc := newTestService()
	_, product, po := setupConfirmedPO(t, svc)
	inbound, _ := svc.CreateInboundOrder(po.ID)
	ins, _ := svc.StartInspection(inbound.ID, "qa")
	_, _ = svc.CompleteInspection(ins.ID, model.InspectionPassed, 0, "ok")
	_, _ = svc.StockInbound(inbound.ID)

	if _, err := svc.CreateReturnOrder(inbound.ID, product.ID, 6, "first"); err != nil {
		t.Fatalf("create first return: %v", err)
	}
	if _, err := svc.CreateReturnOrder(inbound.ID, product.ID, 6, "second"); err == nil {
		t.Fatal("second pending return exceeded inbound quantity")
	}
}
