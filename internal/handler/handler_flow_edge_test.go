package handler

import (
	"net/http"
	"testing"

	"supplychain/internal/config"
)

// 创建供应商+商品+已确认采购单的快捷路径。
func setupConfirmedPO(t *testing.T, tsURL string) (supplierID, productID, poID string) {
	t.Helper()
	_, sb := doJSON(t, http.MethodPost, tsURL+"/api/suppliers", `{"name":"华东电子","phone":"13800000001"}`)
	supplierID = dataID(sb)
	_, pb := doJSON(t, http.MethodPost, tsURL+"/api/products", `{"sku":"SKU001","name":"螺丝","price":100,"low_stock_threshold":10}`)
	productID = dataID(pb)
	_, poBody := doJSON(t, http.MethodPost, tsURL+"/api/purchase-orders",
		`{"supplier_id":"`+supplierID+`","items":[{"product_id":"`+productID+`","quantity":10,"unit_price":100}]}`)
	poID = dataID(poBody)
	doJSON(t, http.MethodPost, tsURL+"/api/purchase-orders/"+poID+"/confirm", "")
	return
}

func TestRejectedInspectionFlow(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()
	_, productID, poID := setupConfirmedPO(t, ts.URL)

	_, inBody := doJSON(t, http.MethodPost, ts.URL+"/api/inbound-orders", `{"purchase_order_id":"`+poID+`"}`)
	inboundID := dataID(inBody)
	_, insBody := doJSON(t, http.MethodPost, ts.URL+"/api/inspections",
		`{"inbound_order_id":"`+inboundID+`","inspector":"质检员"}`)
	insID := dataID(insBody)

	// 质检不合格
	resp, _ := doJSON(t, http.MethodPost, ts.URL+"/api/inspections/"+insID+"/complete",
		`{"result":"failed","defect_count":5,"note":"不合格"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("complete failed status = %d", resp.StatusCode)
	}

	// 不合格不能入库 → 400
	resp, _ = doJSON(t, http.MethodPost, ts.URL+"/api/inbound-orders/"+inboundID+"/stock", "")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("stock failed inspection status = %d, want 400", resp.StatusCode)
	}

	// 库存仍为 0
	_, stockBody := doJSON(t, http.MethodGet, ts.URL+"/api/stock/"+productID, "")
	if stockBody["data"].(map[string]interface{})["quantity"].(float64) != 0 {
		t.Fatal("expected stock 0 after failed inspection")
	}
}

func TestCancelledPOFlow(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()
	_, _, poID := setupConfirmedPO(t, ts.URL)

	// 已确认的采购单不可取消 → 400
	resp, _ := doJSON(t, http.MethodPost, ts.URL+"/api/purchase-orders/"+poID+"/cancel", "")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("cancel confirmed status = %d, want 400", resp.StatusCode)
	}
}

func TestReturnOrderHandlerFlow(t *testing.T) {
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

	// 创建退货单
	resp, body := doJSON(t, http.MethodPost, ts.URL+"/api/return-orders",
		`{"inbound_order_id":"`+inboundID+`","product_id":"`+productID+`","quantity":3,"reason":"质量问题"}`)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create return status = %d", resp.StatusCode)
	}
	returnID := dataID(body)

	// 完成退货
	resp, _ = doJSON(t, http.MethodPost, ts.URL+"/api/return-orders/"+returnID+"/complete", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("complete return status = %d", resp.StatusCode)
	}

	// 库存应减少到 7
	_, stockBody := doJSON(t, http.MethodGet, ts.URL+"/api/stock/"+productID, "")
	if stockBody["data"].(map[string]interface{})["quantity"].(float64) != 7 {
		t.Fatalf("stock after return = %v", stockBody["data"])
	}

	// 退货统计
	resp, _ = doJSON(t, http.MethodGet, ts.URL+"/api/reports/returns", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("return stats status = %d", resp.StatusCode)
	}
}

func TestTraceBatchHandler(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()
	_, _, poID := setupConfirmedPO(t, ts.URL)

	_, inBody := doJSON(t, http.MethodPost, ts.URL+"/api/inbound-orders", `{"purchase_order_id":"`+poID+`"}`)
	inboundID := dataID(inBody)
	_, insBody := doJSON(t, http.MethodPost, ts.URL+"/api/inspections",
		`{"inbound_order_id":"`+inboundID+`","inspector":"质检员"}`)
	insID := dataID(insBody)
	doJSON(t, http.MethodPost, ts.URL+"/api/inspections/"+insID+"/complete", `{"result":"passed"}`)
	doJSON(t, http.MethodPost, ts.URL+"/api/inbound-orders/"+inboundID+"/stock", "")

	// 获取批次 ID
	_, batchBody := doJSON(t, http.MethodGet, ts.URL+"/api/batches", "")
	batchItems := batchBody["data"].(map[string]interface{})["items"].([]interface{})
	if len(batchItems) != 1 {
		t.Fatalf("batches = %d", len(batchItems))
	}
	batchID := batchItems[0].(map[string]interface{})["id"].(string)

	// 追溯
	resp, body := doJSON(t, http.MethodGet, ts.URL+"/api/batches/"+batchID+"/trace", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("trace status = %d", resp.StatusCode)
	}
	trace := body["data"].(map[string]interface{})
	if trace["supplier_name"] != "华东电子" {
		t.Fatalf("trace = %+v", trace)
	}
}

func TestRateLimitHandler(t *testing.T) {
	ts := newTestServerWithConfig(t, &config.Config{MaxPageSize: 100, RateLimitPerSec: 2})
	defer ts.Close()

	// 前 2 个请求放行，第 3 个限流
	for i := 0; i < 2; i++ {
		resp, _ := doJSON(t, http.MethodGet, ts.URL+"/api/suppliers", "")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("request %d status = %d", i, resp.StatusCode)
		}
	}
	resp, _ := doJSON(t, http.MethodGet, ts.URL+"/api/suppliers", "")
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("rate limited status = %d, want 429", resp.StatusCode)
	}
}

func TestListFiltersHandler(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	// 创建多个供应商与商品
	doJSON(t, http.MethodPost, ts.URL+"/api/suppliers", `{"name":"供应商A","phone":"13800000001","status":"active"}`)
	doJSON(t, http.MethodPost, ts.URL+"/api/suppliers", `{"name":"供应商B","phone":"13800000002","status":"inactive"}`)

	_, body := doJSON(t, http.MethodGet, ts.URL+"/api/suppliers?status=active", "")
	if body["data"].(map[string]interface{})["pagination"].(map[string]interface{})["total"].(float64) != 1 {
		t.Fatal("expected 1 active supplier")
	}

	_, body = doJSON(t, http.MethodGet, ts.URL+"/api/suppliers?keyword=供应商B", "")
	if body["data"].(map[string]interface{})["pagination"].(map[string]interface{})["total"].(float64) != 1 {
		t.Fatal("expected 1 supplier matching keyword")
	}
}
