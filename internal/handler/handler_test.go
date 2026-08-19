package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"supplychain/internal/config"
	"supplychain/internal/service"
	"supplychain/internal/store"
	"supplychain/pkg/logger"
)

func newTestServerWithConfig(t *testing.T, cfg *config.Config) *httptest.Server {
	t.Helper()
	if cfg == nil {
		cfg = &config.Config{MaxPageSize: 100}
	}
	log := logger.NewLevel(logger.LevelError)
	svc := service.New(store.NewMemoryStore(), log, cfg)
	srv := NewServer(svc, log, cfg)
	return httptest.NewServer(srv.Routes())
}

func newTestServer(t *testing.T) *httptest.Server {
	return newTestServerWithConfig(t, &config.Config{MaxPageSize: 100})
}

func doJSON(t *testing.T, method, url, body string) (*http.Response, map[string]interface{}) {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = bytes.NewReader([]byte(body))
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var decoded map[string]interface{}
	_ = json.Unmarshal(raw, &decoded)
	return resp, decoded
}

func doJSONAuth(t *testing.T, method, url, body, token string) (*http.Response, map[string]interface{}) {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = bytes.NewReader([]byte(body))
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var decoded map[string]interface{}
	_ = json.Unmarshal(raw, &decoded)
	return resp, decoded
}

func dataID(body map[string]interface{}) string {
	return body["data"].(map[string]interface{})["id"].(string)
}

func TestFullSupplyChainFlow(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	// 供应商
	_, sb := doJSON(t, http.MethodPost, ts.URL+"/api/suppliers",
		`{"name":"华东电子","contact":"王总","phone":"13800000001","address":"上海"}`)
	supplierID := dataID(sb)

	// 商品
	_, pb := doJSON(t, http.MethodPost, ts.URL+"/api/products",
		`{"sku":"SKU001","name":"螺丝","category":"原料","price":100,"low_stock_threshold":10}`)
	productID := dataID(pb)

	// 采购单
	_, poBody := doJSON(t, http.MethodPost, ts.URL+"/api/purchase-orders",
		`{"supplier_id":"`+supplierID+`","items":[{"product_id":"`+productID+`","quantity":10,"unit_price":100}]}`)
	poID := dataID(poBody)

	// 确认采购单
	resp, _ := doJSON(t, http.MethodPost, ts.URL+"/api/purchase-orders/"+poID+"/confirm", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("confirm status = %d", resp.StatusCode)
	}

	// 创建入库单
	_, inBody := doJSON(t, http.MethodPost, ts.URL+"/api/inbound-orders",
		`{"purchase_order_id":"`+poID+`"}`)
	inboundID := dataID(inBody)

	// 开始质检
	_, insBody := doJSON(t, http.MethodPost, ts.URL+"/api/inspections",
		`{"inbound_order_id":"`+inboundID+`","inspector":"质检员"}`)
	insID := dataID(insBody)

	// 完成质检
	resp, _ = doJSON(t, http.MethodPost, ts.URL+"/api/inspections/"+insID+"/complete",
		`{"result":"passed","defect_count":0,"note":"合格"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("complete status = %d", resp.StatusCode)
	}

	// 入库
	resp, _ = doJSON(t, http.MethodPost, ts.URL+"/api/inbound-orders/"+inboundID+"/stock", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("stock status = %d", resp.StatusCode)
	}

	// 查询库存
	resp, stockBody := doJSON(t, http.MethodGet, ts.URL+"/api/stock/"+productID, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("stock status = %d", resp.StatusCode)
	}
	if stockBody["data"].(map[string]interface{})["quantity"].(float64) != 10 {
		t.Fatalf("stock = %v", stockBody["data"])
	}

	// 出库
	resp, _ = doJSON(t, http.MethodPost, ts.URL+"/api/stock/outbound",
		`{"product_id":"`+productID+`","quantity":4,"note":"销售出库"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("outbound status = %d", resp.StatusCode)
	}

	// 再查库存
	_, stockBody = doJSON(t, http.MethodGet, ts.URL+"/api/stock/"+productID, "")
	if stockBody["data"].(map[string]interface{})["quantity"].(float64) != 6 {
		t.Fatalf("stock after outbound = %v", stockBody["data"])
	}
}

func TestAuthRequired(t *testing.T) {
	ts := newTestServerWithConfig(t, &config.Config{MaxPageSize: 100, AuthToken: "secret"})
	defer ts.Close()

	// 未携带 token → 401
	resp, _ := doJSON(t, http.MethodGet, ts.URL+"/api/suppliers", "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("no token status = %d, want 401", resp.StatusCode)
	}

	// 携带正确 token → 200
	resp, _ = doJSONAuth(t, http.MethodGet, ts.URL+"/api/suppliers", "", "secret")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("with token status = %d, want 200", resp.StatusCode)
	}

	// 错误 token → 401
	resp, _ = doJSONAuth(t, http.MethodGet, ts.URL+"/api/suppliers", "", "wrong")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("wrong token status = %d, want 401", resp.StatusCode)
	}
}

func TestErrorCases(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	// 404
	resp, _ := doJSON(t, http.MethodGet, ts.URL+"/api/suppliers/missing", "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("missing status = %d, want 404", resp.StatusCode)
	}

	// 校验失败 400
	resp, _ = doJSON(t, http.MethodPost, ts.URL+"/api/suppliers", `{"name":"","phone":"13800000001"}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid status = %d, want 400", resp.StatusCode)
	}

	// 畸形 JSON 400
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/products", bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("malformed status = %d, want 400", resp.StatusCode)
	}
}

func TestConflictScenarios(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	doJSON(t, http.MethodPost, ts.URL+"/api/products", `{"sku":"SKU001","name":"螺丝"}`)
	// 重复 SKU → 409
	resp, _ := doJSON(t, http.MethodPost, ts.URL+"/api/products", `{"sku":"SKU001","name":"另一螺丝"}`)
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate sku status = %d, want 409", resp.StatusCode)
	}
}

func TestReportsEndpoints(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	_, sb := doJSON(t, http.MethodPost, ts.URL+"/api/suppliers",
		`{"name":"华东电子","phone":"13800000001"}`)
	supplierID := dataID(sb)
	_, pb := doJSON(t, http.MethodPost, ts.URL+"/api/products",
		`{"sku":"SKU001","name":"螺丝","price":100,"low_stock_threshold":10}`)
	productID := dataID(pb)
	_, poBody := doJSON(t, http.MethodPost, ts.URL+"/api/purchase-orders",
		`{"supplier_id":"`+supplierID+`","items":[{"product_id":"`+productID+`","quantity":10,"unit_price":100}]}`)
	poID := dataID(poBody)
	doJSON(t, http.MethodPost, ts.URL+"/api/purchase-orders/"+poID+"/confirm", "")
	_, inBody := doJSON(t, http.MethodPost, ts.URL+"/api/inbound-orders",
		`{"purchase_order_id":"`+poID+`"}`)
	inboundID := dataID(inBody)
	_, insBody := doJSON(t, http.MethodPost, ts.URL+"/api/inspections",
		`{"inbound_order_id":"`+inboundID+`","inspector":"质检员"}`)
	insID := dataID(insBody)
	doJSON(t, http.MethodPost, ts.URL+"/api/inspections/"+insID+"/complete", `{"result":"passed"}`)
	doJSON(t, http.MethodPost, ts.URL+"/api/inbound-orders/"+inboundID+"/stock", "")

	for _, path := range []string{
		"/api/reports/stock",
		"/api/reports/quality",
		"/api/reports/supplier-ranking",
		"/api/reports/inventory-value",
	} {
		resp, _ := doJSON(t, http.MethodGet, ts.URL+path, "")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s status = %d", path, resp.StatusCode)
		}
	}
}

func TestImportExportEndpoints(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, body := doJSON(t, http.MethodPost, ts.URL+"/api/products/import",
		`{"products":[{"sku":"SKU001","name":"螺丝","price":100},{"sku":"SKU002","name":"螺母","price":200}]}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("import status = %d", resp.StatusCode)
	}
	if len(body["data"].([]interface{})) != 2 {
		t.Fatalf("import results = %d", len(body["data"].([]interface{})))
	}

	resp, _ = doJSON(t, http.MethodGet, ts.URL+"/api/export", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("export status = %d", resp.StatusCode)
	}
}
