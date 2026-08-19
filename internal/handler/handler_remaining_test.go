package handler

import (
	"net/http"
	"testing"
)

func TestGetMissingEntities(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	paths := []string{
		"/api/suppliers/missing",
		"/api/products/missing",
		"/api/purchase-orders/missing",
		"/api/inbound-orders/missing",
		"/api/inspections/missing",
		"/api/return-orders/missing",
	}
	for _, path := range paths {
		resp, _ := doJSON(t, http.MethodGet, ts.URL+path, "")
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("%s status = %d, want 404", path, resp.StatusCode)
		}
	}
}

func TestDeleteEntitiesHandler(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	// 供应商删除
	_, sb := doJSON(t, http.MethodPost, ts.URL+"/api/suppliers", `{"name":"华东电子","phone":"13800000001"}`)
	supplierID := dataID(sb)
	resp, _ := doJSON(t, http.MethodDelete, ts.URL+"/api/suppliers/"+supplierID, "")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete supplier status = %d", resp.StatusCode)
	}

	// 商品删除
	_, pb := doJSON(t, http.MethodPost, ts.URL+"/api/products", `{"sku":"SKU001","name":"螺丝"}`)
	productID := dataID(pb)
	resp, _ = doJSON(t, http.MethodDelete, ts.URL+"/api/products/"+productID, "")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete product status = %d", resp.StatusCode)
	}

	// 删除不存在的 → 404
	resp, _ = doJSON(t, http.MethodDelete, ts.URL+"/api/suppliers/missing", "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("delete missing status = %d, want 404", resp.StatusCode)
	}
}

func TestStartInspectionValidation(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	// 入库单不存在 → 404
	resp, _ := doJSON(t, http.MethodPost, ts.URL+"/api/inspections",
		`{"inbound_order_id":"missing","inspector":"质检员"}`)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("start inspection missing status = %d, want 404", resp.StatusCode)
	}

	// 缺质检员 → 400
	_, productID, poID := setupConfirmedPO(t, ts.URL)
	_, inBody := doJSON(t, http.MethodPost, ts.URL+"/api/inbound-orders", `{"purchase_order_id":"`+poID+`"}`)
	inboundID := dataID(inBody)
	resp, _ = doJSON(t, http.MethodPost, ts.URL+"/api/inspections",
		`{"inbound_order_id":"`+inboundID+`","inspector":""}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("empty inspector status = %d, want 400", resp.StatusCode)
	}
	_ = productID
}

func TestStockInboundValidation(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	// 入库单不存在 → 404
	resp, _ := doJSON(t, http.MethodPost, ts.URL+"/api/inbound-orders/missing/stock", "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("stock missing status = %d, want 404", resp.StatusCode)
	}

	// 待入库状态直接入库 → 400
	_, _, poID := setupConfirmedPO(t, ts.URL)
	_, inBody := doJSON(t, http.MethodPost, ts.URL+"/api/inbound-orders", `{"purchase_order_id":"`+poID+`"}`)
	inboundID := dataID(inBody)
	resp, _ = doJSON(t, http.MethodPost, ts.URL+"/api/inbound-orders/"+inboundID+"/stock", "")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("stock pending status = %d, want 400", resp.StatusCode)
	}
}

func TestOutboundValidation(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	_, pb := doJSON(t, http.MethodPost, ts.URL+"/api/products", `{"sku":"SKU001","name":"螺丝"}`)
	productID := dataID(pb)

	// 库存不足 → 400
	resp, _ := doJSON(t, http.MethodPost, ts.URL+"/api/stock/outbound",
		`{"product_id":"`+productID+`","quantity":10,"note":"出库"}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("outbound empty stock status = %d, want 400", resp.StatusCode)
	}
}
