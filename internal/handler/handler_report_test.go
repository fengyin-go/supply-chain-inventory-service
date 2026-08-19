package handler

import (
	"net/http"
	"testing"
)

func TestReportEndpointsExtended(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	// 建立一条完整入库链路
	_, productID, poID := setupConfirmedPO(t, ts.URL)
	_, inBody := doJSON(t, http.MethodPost, ts.URL+"/api/inbound-orders", `{"purchase_order_id":"`+poID+`"}`)
	inboundID := dataID(inBody)
	_, insBody := doJSON(t, http.MethodPost, ts.URL+"/api/inspections",
		`{"inbound_order_id":"`+inboundID+`","inspector":"质检员"}`)
	insID := dataID(insBody)
	doJSON(t, http.MethodPost, ts.URL+"/api/inspections/"+insID+"/complete", `{"result":"passed"}`)
	doJSON(t, http.MethodPost, ts.URL+"/api/inbound-orders/"+inboundID+"/stock", "")
	doJSON(t, http.MethodPost, ts.URL+"/api/stock/outbound", `{"product_id":"`+productID+`","quantity":3,"note":"出库"}`)

	paths := []string{
		"/api/reports/purchase-summary",
		"/api/reports/movements",
		"/api/reports/categories",
	}
	for _, path := range paths {
		resp, _ := doJSON(t, http.MethodGet, ts.URL+path, "")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s status = %d", path, resp.StatusCode)
		}
	}
}

func TestMovementsFilterHandler(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	_, productID, poID := setupConfirmedPO(t, ts.URL)
	_, inBody := doJSON(t, http.MethodPost, ts.URL+"/api/inbound-orders", `{"purchase_order_id":"`+poID+`"}`)
	inboundID := dataID(inBody)
	_, insBody := doJSON(t, http.MethodPost, ts.URL+"/api/inspections",
		`{"inbound_order_id":"`+inboundID+`","inspector":"质检员"}`)
	insID := dataID(insBody)
	doJSON(t, http.MethodPost, ts.URL+"/api/inspections/"+insID+"/complete", `{"result":"passed"}`)
	doJSON(t, http.MethodPost, ts.URL+"/api/inbound-orders/"+inboundID+"/stock", "")
	doJSON(t, http.MethodPost, ts.URL+"/api/stock/outbound", `{"product_id":"`+productID+`","quantity":2,"note":"出库"}`)

	// 按类型筛选出库流水
	_, body := doJSON(t, http.MethodGet, ts.URL+"/api/movements?type=outbound", "")
	if body["data"].(map[string]interface{})["pagination"].(map[string]interface{})["total"].(float64) != 1 {
		t.Fatal("expected 1 outbound movement")
	}

	// 按商品筛选
	_, body = doJSON(t, http.MethodGet, ts.URL+"/api/movements?product_id="+productID, "")
	if body["data"].(map[string]interface{})["pagination"].(map[string]interface{})["total"].(float64) != 2 {
		t.Fatal("expected 2 movements for product")
	}
}

func TestBatchFilterHandler(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	_, productID, poID := setupConfirmedPO(t, ts.URL)
	_, inBody := doJSON(t, http.MethodPost, ts.URL+"/api/inbound-orders", `{"purchase_order_id":"`+poID+`"}`)
	inboundID := dataID(inBody)
	_, insBody := doJSON(t, http.MethodPost, ts.URL+"/api/inspections",
		`{"inbound_order_id":"`+inboundID+`","inspector":"质检员"}`)
	insID := dataID(insBody)
	doJSON(t, http.MethodPost, ts.URL+"/api/inspections/"+insID+"/complete", `{"result":"passed"}`)
	doJSON(t, http.MethodPost, ts.URL+"/api/inbound-orders/"+inboundID+"/stock", "")

	// 按商品筛选批次
	_, body := doJSON(t, http.MethodGet, ts.URL+"/api/batches?product_id="+productID, "")
	if body["data"].(map[string]interface{})["pagination"].(map[string]interface{})["total"].(float64) != 1 {
		t.Fatal("expected 1 batch")
	}
}

func TestLowStockReportHandler(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	// 商品库存 0，阈值 10 → 低库存
	doJSON(t, http.MethodPost, ts.URL+"/api/products",
		`{"sku":"SKU001","name":"螺丝","price":100,"low_stock_threshold":10}`)

	resp, body := doJSON(t, http.MethodGet, ts.URL+"/api/reports/stock?low=true", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("low stock status = %d", resp.StatusCode)
	}
	list := body["data"].([]interface{})
	if len(list) != 1 {
		t.Fatalf("low stock len = %d, want 1", len(list))
	}
	if list[0].(map[string]interface{})["low_stock"] != true {
		t.Fatal("expected low_stock true")
	}
}
