package handler

import (
	"net/http"
	"testing"
)

func TestSupplierCRUDHandler(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	// 创建
	resp, body := doJSON(t, http.MethodPost, ts.URL+"/api/suppliers",
		`{"name":"华东电子","contact":"王总","phone":"13800000001","address":"上海"}`)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d", resp.StatusCode)
	}
	id := dataID(body)

	// 查询
	resp, _ = doJSON(t, http.MethodGet, ts.URL+"/api/suppliers/"+id, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get status = %d", resp.StatusCode)
	}

	// 更新
	resp, body = doJSON(t, http.MethodPut, ts.URL+"/api/suppliers/"+id,
		`{"name":"华东电子科技","contact":"李总","phone":"13800000002","address":"苏州","status":"inactive"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update status = %d", resp.StatusCode)
	}
	if body["data"].(map[string]interface{})["status"] != "inactive" {
		t.Fatal("expected inactive status")
	}

	// 按状态筛选
	resp, body = doJSON(t, http.MethodGet, ts.URL+"/api/suppliers?status=inactive", "")
	if body["data"].(map[string]interface{})["pagination"].(map[string]interface{})["total"].(float64) != 1 {
		t.Fatal("expected 1 inactive supplier")
	}

	// 删除
	resp, _ = doJSON(t, http.MethodDelete, ts.URL+"/api/suppliers/"+id, "")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d", resp.StatusCode)
	}
}

func TestProductCRUDHandler(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	resp, body := doJSON(t, http.MethodPost, ts.URL+"/api/products",
		`{"sku":"SKU001","name":"螺丝","category":"原料","price":100,"low_stock_threshold":10}`)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d", resp.StatusCode)
	}
	id := dataID(body)

	resp, _ = doJSON(t, http.MethodGet, ts.URL+"/api/products/"+id, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get status = %d", resp.StatusCode)
	}

	resp, _ = doJSON(t, http.MethodPut, ts.URL+"/api/products/"+id,
		`{"sku":"SKU001","name":"不锈钢螺丝","category":"紧固件","price":150,"low_stock_threshold":20}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update status = %d", resp.StatusCode)
	}

	resp, body = doJSON(t, http.MethodGet, ts.URL+"/api/products?category=紧固件", "")
	if body["data"].(map[string]interface{})["pagination"].(map[string]interface{})["total"].(float64) != 1 {
		t.Fatal("expected 1 product in category")
	}

	resp, _ = doJSON(t, http.MethodDelete, ts.URL+"/api/products/"+id, "")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d", resp.StatusCode)
	}
}

func TestPurchaseOrderCRUDHandler(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	_, sb := doJSON(t, http.MethodPost, ts.URL+"/api/suppliers", `{"name":"华东电子","phone":"13800000001"}`)
	supplierID := dataID(sb)
	_, pb := doJSON(t, http.MethodPost, ts.URL+"/api/products", `{"sku":"SKU001","name":"螺丝","price":100}`)
	productID := dataID(pb)

	resp, body := doJSON(t, http.MethodPost, ts.URL+"/api/purchase-orders",
		`{"supplier_id":"`+supplierID+`","items":[{"product_id":"`+productID+`","quantity":5,"unit_price":100}]}`)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d", resp.StatusCode)
	}
	id := dataID(body)

	resp, _ = doJSON(t, http.MethodGet, ts.URL+"/api/purchase-orders/"+id, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get status = %d", resp.StatusCode)
	}

	resp, body = doJSON(t, http.MethodGet, ts.URL+"/api/purchase-orders?status=draft", "")
	if body["data"].(map[string]interface{})["pagination"].(map[string]interface{})["total"].(float64) != 1 {
		t.Fatal("expected 1 draft order")
	}

	// 取消
	resp, _ = doJSON(t, http.MethodPost, ts.URL+"/api/purchase-orders/"+id+"/cancel", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("cancel status = %d", resp.StatusCode)
	}
}

func TestInboundAndInspectionListHandler(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	_, sb := doJSON(t, http.MethodPost, ts.URL+"/api/suppliers", `{"name":"华东电子","phone":"13800000001"}`)
	supplierID := dataID(sb)
	_, pb := doJSON(t, http.MethodPost, ts.URL+"/api/products", `{"sku":"SKU001","name":"螺丝","price":100}`)
	productID := dataID(pb)
	_, poBody := doJSON(t, http.MethodPost, ts.URL+"/api/purchase-orders",
		`{"supplier_id":"`+supplierID+`","items":[{"product_id":"`+productID+`","quantity":5,"unit_price":100}]}`)
	poID := dataID(poBody)
	doJSON(t, http.MethodPost, ts.URL+"/api/purchase-orders/"+poID+"/confirm", "")
	_, inBody := doJSON(t, http.MethodPost, ts.URL+"/api/inbound-orders",
		`{"purchase_order_id":"`+poID+`"}`)
	inboundID := dataID(inBody)
	_, insBody := doJSON(t, http.MethodPost, ts.URL+"/api/inspections",
		`{"inbound_order_id":"`+inboundID+`","inspector":"质检员"}`)
	insID := dataID(insBody)

	// 入库单列表
	resp, body := doJSON(t, http.MethodGet, ts.URL+"/api/inbound-orders", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list inbound status = %d", resp.StatusCode)
	}
	if body["data"].(map[string]interface{})["pagination"].(map[string]interface{})["total"].(float64) != 1 {
		t.Fatal("expected 1 inbound")
	}

	// 按采购单筛选
	resp, _ = doJSON(t, http.MethodGet, ts.URL+"/api/inbound-orders?purchase_order_id="+poID, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("filter inbound status = %d", resp.StatusCode)
	}

	// 质检列表
	resp, body = doJSON(t, http.MethodGet, ts.URL+"/api/inspections", "")
	if body["data"].(map[string]interface{})["pagination"].(map[string]interface{})["total"].(float64) != 1 {
		t.Fatal("expected 1 inspection")
	}

	// 查询质检详情
	resp, _ = doJSON(t, http.MethodGet, ts.URL+"/api/inspections/"+insID, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get inspection status = %d", resp.StatusCode)
	}
}

func TestBatchesAndMovementsListHandler(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	_, pb := doJSON(t, http.MethodPost, ts.URL+"/api/products",
		`{"sku":"SKU001","name":"螺丝","price":100}`)
	productID := dataID(pb)

	// 盘点调整增加库存
	doJSON(t, http.MethodPost, ts.URL+"/api/stock/adjust",
		`{"product_id":"`+productID+`","delta":20,"note":"期初"}`)

	resp, body := doJSON(t, http.MethodGet, ts.URL+"/api/batches", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("batches status = %d", resp.StatusCode)
	}
	if body["data"].(map[string]interface{})["pagination"].(map[string]interface{})["total"].(float64) != 1 {
		t.Fatal("expected 1 batch")
	}

	resp, body = doJSON(t, http.MethodGet, ts.URL+"/api/movements", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("movements status = %d", resp.StatusCode)
	}
	if body["data"].(map[string]interface{})["pagination"].(map[string]interface{})["total"].(float64) != 1 {
		t.Fatal("expected 1 movement")
	}
}

func TestPaginationHandler(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	for i := 0; i < 5; i++ {
		doJSON(t, http.MethodPost, ts.URL+"/api/suppliers",
			`{"name":"供应商`+string(rune('A'+i))+`","phone":"1380000000`+string(rune('1'+i))+`"}`)
	}

	resp, body := doJSON(t, http.MethodGet, ts.URL+"/api/suppliers?page=1&size=2", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d", resp.StatusCode)
	}
	pagination := body["data"].(map[string]interface{})["pagination"].(map[string]interface{})
	if pagination["total"].(float64) != 5 || pagination["size"].(float64) != 2 {
		t.Fatalf("pagination = %+v", pagination)
	}
	items := body["data"].(map[string]interface{})["items"].([]interface{})
	if len(items) != 2 {
		t.Fatalf("page items = %d, want 2", len(items))
	}

	// 越界页
	resp, body = doJSON(t, http.MethodGet, ts.URL+"/api/suppliers?page=99&size=2", "")
	items = body["data"].(map[string]interface{})["items"].([]interface{})
	if len(items) != 0 {
		t.Fatalf("empty page items = %d", len(items))
	}
}
