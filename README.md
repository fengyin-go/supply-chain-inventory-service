# 供应链管理系统（supplychain）

纯 Go 标准库实现的供应链后端，零第三方依赖。**金额字段统一使用「分」为单位（int64）。**

## 运行

```bash
go run ./cmd/server
```

环境变量：

| 变量 | 默认 | 说明 |
|------|------|------|
| PORT / ADDR | :8080 | 监听地址 |
| MAX_PAGE_SIZE | 100 | 分页最大条数 |
| AUTH_TOKEN | 空 | 非空时启用 Bearer Token 鉴权 |
| RATE_LIMIT | 0 | 每 IP 每秒限流数，0 表示不限流 |
| LOG_LEVEL | info | 日志级别（debug/info/warn/error） |

## 核心业务流程

1. 创建供应商、商品。
2. 创建采购单（草稿）→ 确认 → 创建入库单 → 开始质检 → 质检合格 → 入库（生成批次与库存）。
3. 出库按先进先出扣减批次库存。
4. 支持采购退货、盘点调整、库存流水与批次追溯。

状态机：

- 采购单：`draft → confirmed → received`，`draft → cancelled`
- 入库单：`pending → inspecting → stocked / rejected`
- 质检：`pending → passed / failed`
- 退货单：`pending → completed`

## API

### 供应商

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/suppliers | 创建供应商 |
| GET | /api/suppliers?status=&keyword=&page=&size= | 分页查询 |
| GET | /api/suppliers/{id} | 查询单个 |
| PUT | /api/suppliers/{id} | 更新 |
| DELETE | /api/suppliers/{id} | 删除 |

### 商品

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/products | 创建商品 |
| GET | /api/products?category=&keyword=&page=&size= | 分页查询 |
| GET | /api/products/{id} | 查询单个 |
| PUT | /api/products/{id} | 更新 |
| DELETE | /api/products/{id} | 删除 |
| POST | /api/products/import | 批量导入商品（部分成功不阻断） |

### 采购单

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/purchase-orders | 创建采购单（草稿） |
| GET | /api/purchase-orders?status=&supplier_id=&page=&size= | 分页查询 |
| GET | /api/purchase-orders/{id} | 查询单个 |
| POST | /api/purchase-orders/{id}/confirm | 确认 |
| POST | /api/purchase-orders/{id}/cancel | 取消 |
| DELETE | /api/purchase-orders/{id} | 删除 |

### 入库单

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/inbound-orders | 从已确认采购单创建入库单 |
| GET | /api/inbound-orders?status=&purchase_order_id=&page=&size= | 分页查询 |
| GET | /api/inbound-orders/{id} | 查询单个 |
| POST | /api/inbound-orders/{id}/reject | 驳回 |
| POST | /api/inbound-orders/{id}/stock | 入库（质检合格后） |
| DELETE | /api/inbound-orders/{id} | 删除 |

### 质检

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/inspections | 开始质检（body: `{"inbound_order_id","inspector"}`） |
| GET | /api/inspections?result=&inbound_order_id=&page=&size= | 分页查询 |
| GET | /api/inspections/{id} | 查询单个 |
| POST | /api/inspections/{id}/complete | 完成质检（passed/failed） |

### 库存

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/stock/{product_id} | 查询商品库存 |
| POST | /api/stock/outbound | 出库（先进先出） |
| POST | /api/stock/adjust | 盘点调整（正数增加/负数扣减） |
| GET | /api/batches?product_id=&page=&size= | 分页查询批次 |
| GET | /api/batches/{id}/trace | 批次追溯 |
| GET | /api/movements?product_id=&type=&page=&size= | 分页查询库存流水 |

### 退货单

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/return-orders | 创建退货单 |
| GET | /api/return-orders?status=&inbound_order_id=&page=&size= | 分页查询 |
| GET | /api/return-orders/{id} | 查询单个 |
| POST | /api/return-orders/{id}/complete | 完成退货（扣减库存） |

### 报表

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/reports/stock?low=true | 库存报表（可只看低库存） |
| GET | /api/reports/quality | 质检合格率 |
| GET | /api/reports/supplier-ranking?limit= | 供应商采购额排行 |
| GET | /api/reports/purchase-summary | 采购汇总 |
| GET | /api/reports/inventory-value | 库存总金额 |
| GET | /api/reports/movements | 库存流水统计 |
| GET | /api/reports/categories | 商品分类统计 |
| GET | /api/reports/returns | 退货统计 |

### 导出

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/export | 全量数据导出 |

## 分层

- `cmd/server` 入口；`internal/app` 依赖装配；`internal/config` 环境变量配置。
- `internal/middleware` 鉴权、限流中间件。
- `internal/model` 领域模型、校验与状态机；`internal/store` 内存存储；`internal/service` 业务逻辑；`internal/handler` HTTP 路由。
- `pkg/httpx`、`pkg/idgen`、`pkg/logger` 通用工具。
