package handler

import (
	"net/http"
	"testing"
)

func TestImportProductsPartialFailure(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	// 一个合法，一个缺 SKU
	resp, body := doJSON(t, http.MethodPost, ts.URL+"/api/products/import",
		`{"products":[{"sku":"SKU001","name":"螺丝","price":100},{"sku":"","name":"缺SKU","price":100}]}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("import status = %d", resp.StatusCode)
	}
	items := body["data"].([]interface{})
	if len(items) != 2 {
		t.Fatalf("import results = %d", len(items))
	}
	if items[0].(map[string]interface{})["success"] != true {
		t.Fatal("first should succeed")
	}
	if items[1].(map[string]interface{})["success"] != false {
		t.Fatal("second should fail")
	}
}

func TestImportEmptyList(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, _ := doJSON(t, http.MethodPost, ts.URL+"/api/products/import", `{"products":[]}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("empty import status = %d, want 400", resp.StatusCode)
	}
}

func TestExportEmptyStore(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, body := doJSON(t, http.MethodGet, ts.URL+"/api/export", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("export status = %d", resp.StatusCode)
	}
	data := body["data"].(map[string]interface{})
	if len(data["suppliers"].([]interface{})) != 0 {
		t.Fatal("expected empty suppliers")
	}
}

func TestSupplierKeywordSearchHandler(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	doJSON(t, http.MethodPost, ts.URL+"/api/suppliers", `{"name":"华东电子","phone":"13800000001"}`)
	doJSON(t, http.MethodPost, ts.URL+"/api/suppliers", `{"name":"华西机械","phone":"13800000002"}`)

	_, body := doJSON(t, http.MethodGet, ts.URL+"/api/suppliers?keyword=华西", "")
	if body["data"].(map[string]interface{})["pagination"].(map[string]interface{})["total"].(float64) != 1 {
		t.Fatal("expected 1 match for 华西")
	}
}

func TestProductCategorySearchHandler(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	doJSON(t, http.MethodPost, ts.URL+"/api/products", `{"sku":"SKU001","name":"螺丝","category":"紧固件"}`)
	doJSON(t, http.MethodPost, ts.URL+"/api/products", `{"sku":"SKU002","name":"螺母","category":"五金"}`)

	_, body := doJSON(t, http.MethodGet, ts.URL+"/api/products?category=五金", "")
	if body["data"].(map[string]interface{})["pagination"].(map[string]interface{})["total"].(float64) != 1 {
		t.Fatal("expected 1 match for 五金")
	}
}
