package service

import (
	"testing"

	"supplychain/internal/model"
)

func TestInboundSnapshotIgnoresPurchaseOrderItemMutation(t *testing.T) {
	svc := newTestService()

	sup, err := svc.CreateSupplier(model.Supplier{Name: "North Dock", Phone: "13800000000"})
	if err != nil {
		t.Fatalf("create supplier: %v", err)
	}
	product, err := svc.CreateProduct(model.Product{SKU: "BOARD-A", Name: "control board", Unit: "pcs", Price: 1200})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}
	po, err := svc.CreatePurchaseOrder(model.PurchaseOrder{
		SupplierID: sup.ID,
		Items: []model.PurchaseItem{{
			ProductID: product.ID,
			Quantity:  5,
			UnitPrice: 1200,
		}},
	})
	if err != nil {
		t.Fatalf("create purchase order: %v", err)
	}
	if _, err := svc.ConfirmPurchaseOrder(po.ID); err != nil {
		t.Fatalf("confirm purchase order: %v", err)
	}
	inbound, err := svc.CreateInboundOrder(po.ID)
	if err != nil {
		t.Fatalf("create inbound order: %v", err)
	}

	loadedPO, err := svc.GetPurchaseOrder(po.ID)
	if err != nil {
		t.Fatalf("load purchase order: %v", err)
	}
	loadedPO.Items[0].Quantity = 50
	loadedPO.Items[0].Amount = 60000

	inspection, err := svc.StartInspection(inbound.ID, "qa")
	if err != nil {
		t.Fatalf("start inspection: %v", err)
	}
	if _, err := svc.CompleteInspection(inspection.ID, model.InspectionPassed, 0, "ok"); err != nil {
		t.Fatalf("complete inspection: %v", err)
	}
	if _, err := svc.StockInbound(inbound.ID); err != nil {
		t.Fatalf("stock inbound: %v", err)
	}
	stock, err := svc.GetStock(product.ID)
	if err != nil {
		t.Fatalf("get stock: %v", err)
	}
	if stock != 5 {
		t.Fatalf("stock after inbound = %d, want original inbound quantity 5", stock)
	}

	if _, err := svc.CreateReturnOrder(inbound.ID, product.ID, 8, "too many"); err == nil {
		t.Fatalf("return above original inbound quantity was accepted")
	}
}
