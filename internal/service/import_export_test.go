package service

import (
	"testing"

	"supplychain/internal/model"
)

func TestImportProducts(t *testing.T) {
	s := newTestService()

	results, err := s.ImportProducts([]model.Product{
		{SKU: "SKU001", Name: "螺丝", Price: 100, LowStockThreshold: 10},
		{SKU: "", Name: "缺SKU", Price: 100},
		{SKU: "SKU002", Name: "螺母", Price: 200, LowStockThreshold: 5},
	})
	if err != nil {
		t.Fatalf("ImportProducts: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("len = %d, want 3", len(results))
	}
	if !results[0].Success || results[1].Success || !results[2].Success {
		t.Fatalf("results = %+v", results)
	}
}

func TestExportSnapshot(t *testing.T) {
	s := newTestService()
	seedSupplier(t, s, "华东电子", "13800000001")
	seedProduct(t, s, "SKU001", "螺丝", 100, 10)

	snapshot, err := s.Export()
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if len(snapshot.Suppliers) != 1 || len(snapshot.Products) != 1 {
		t.Fatalf("snapshot = %+v", snapshot)
	}
	if snapshot.ExportedAt.IsZero() {
		t.Fatal("expected exported_at")
	}
}

func TestGenerateNoUniqueness(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		no := generateNo("PO")
		if seen[no] {
			t.Fatalf("duplicate no: %s", no)
		}
		seen[no] = true
	}
}
