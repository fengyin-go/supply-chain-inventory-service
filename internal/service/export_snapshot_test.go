package service

import (
	"testing"

	"supplychain/internal/model"
)

func TestExportReturnsAnIsolatedSnapshot(t *testing.T) {
	svc := newTestService()
	_, product, po := setupConfirmedPO(t, svc)
	inbound, _ := svc.CreateInboundOrder(po.ID)
	ins, _ := svc.StartInspection(inbound.ID, "qa")
	_, _ = svc.CompleteInspection(ins.ID, model.InspectionPassed, 0, "ok")
	_, _ = svc.StockInbound(inbound.ID)

	snapshot, err := svc.Export()
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	snapshot.Products[0].Name = "tampered"
	snapshot.Products[0].Price = 9999
	snapshot.PurchaseOrders[0].Items[0].Quantity = 99

	stored, _ := svc.GetProduct(product.ID)
	if stored.Name == "tampered" || stored.Price == 9999 {
		t.Fatal("mutating export changed stored product")
	}
	report, _ := svc.StockReport(false)
	if len(report) != 1 || report[0].Value != 1000 {
		t.Fatalf("report after export mutation = %+v, want value 1000", report)
	}
}
