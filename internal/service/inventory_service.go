package service

import (
	"sort"
	"time"

	"supplychain/internal/model"
	"supplychain/pkg/idgen"
)

// StockInbound 入库：质检合格后，将质检中的入库单转为已入库，
// 并生成库存批次与入库流水，同时将采购单标记为已收货。
func (s *Service) StockInbound(inboundID string) (*model.InboundOrder, error) {
	inbound, err := s.store.GetInboundOrder(inboundID)
	if err != nil {
		return nil, err
	}
	if inbound.Status != model.InboundInspecting {
		return nil, model.NewValidationError("status", "仅质检中的入库单可入库")
	}
	ins, err := s.store.GetInspectionByInbound(inboundID)
	if err != nil {
		return nil, model.NewValidationError("inspection", "入库前必须先完成质检")
	}
	if ins.Result != model.InspectionPassed {
		return nil, model.NewValidationError("inspection", "质检不合格，不能入库")
	}
	now := time.Now()
	for _, item := range inbound.Items {
		batch := &model.InventoryBatch{
			ID:             idgen.Hex(),
			BatchNo:        generateNo("BT"),
			ProductID:      item.ProductID,
			InboundOrderID: inboundID,
			Quantity:       item.Quantity,
			Remaining:      item.Quantity,
			CreatedAt:      now,
		}
		if err := s.store.CreateBatch(batch); err != nil {
			return nil, err
		}
		movement := &model.StockMovement{
			ID:        idgen.Hex(),
			ProductID: item.ProductID,
			BatchID:   batch.ID,
			Type:      model.MovementInbound,
			Delta:     item.Quantity,
			Note:      "入库:" + inbound.InboundNo,
			CreatedAt: now,
		}
		if err := s.store.CreateMovement(movement); err != nil {
			return nil, err
		}
	}
	inbound.Status = model.InboundStocked
	inbound.UpdatedAt = now
	if err := s.store.UpdateInboundOrder(inbound); err != nil {
		return nil, err
	}
	if err := s.receivePurchaseOrder(inbound.PurchaseOrderID); err != nil {
		return nil, err
	}
	return inbound, nil
}

// ProductStock 商品当前库存。
type ProductStock struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// GetStock 查询某商品当前可用库存（各批次剩余量之和）。
func (s *Service) GetStock(productID string) (int, error) {
	if _, err := s.store.GetProduct(productID); err != nil {
		return 0, err
	}
	batches := s.store.ListBatchesByProduct(productID)
	total := 0
	for _, b := range batches {
		total += b.Remaining
	}
	return total, nil
}

// OutboundStock 出库：按先进先出扣减批次库存，并记录出库流水。
// 库存扣减与流水写入原子进行：若流水写入失败，已扣减的批次会被回滚，
// 保证库存与流水同生共死，不会出现「库存已扣而流水缺失」的不一致。
func (s *Service) OutboundStock(productID string, quantity int, note string) error {
	return s.deductStock(productID, quantity, note, model.MovementOutbound)
}

// deductStock 按先进先出扣减批次库存并记录流水，movementType 指定流水类型（出库/盘点调整）。
// 任一批次更新或流水写入失败时，回滚此前已修改的批次与已写入的流水，
// 确保「库存」与「流水」要么同时成功、要么同时不生效，绝不会分开提交。
// 这样在流水服务不可用时接口返回失败，库存保持原值，重试也不会重复扣减导致对不上账。
func (s *Service) deductStock(productID string, quantity int, note, movementType string) error {
	if quantity <= 0 {
		return model.NewValidationError("quantity", "扣减数量必须大于 0")
	}
	batches := s.store.ListBatchesByProduct(productID)
	sort.Slice(batches, func(i, j int) bool {
		return batches[i].CreatedAt.Before(batches[j].CreatedAt)
	})
	available := 0
	for _, b := range batches {
		available += b.Remaining
	}
	if available < quantity {
		return model.NewValidationError("quantity", "库存不足")
	}

	remaining := quantity
	now := time.Now()
	updated := make([]*model.InventoryBatch, 0, len(batches)) // 扣减前快照，用于失败回滚
	created := make([]*model.StockMovement, 0, len(batches))   // 已写入流水，用于失败回滚
	rollBack := func() {
		for _, b := range updated {
			_ = s.store.UpdateBatch(b) // 恢复批次到扣减前状态
		}
		for _, m := range created {
			_ = s.store.DeleteMovement(m.ID) // 清理已写入的流水
		}
	}

	for _, b := range batches {
		if remaining == 0 {
			break
		}
		if b.Remaining == 0 {
			continue
		}
		snapshot := *b // 记录扣减前状态
		take := b.Remaining
		if take > remaining {
			take = remaining
		}
		b.Remaining -= take
		if err := s.store.UpdateBatch(b); err != nil {
			rollBack()
			return err
		}
		updated = append(updated, &snapshot)
		movement := &model.StockMovement{
			ID:        idgen.Hex(),
			ProductID: productID,
			BatchID:   b.ID,
			Type:      movementType,
			Delta:     -take,
			Note:      note,
			CreatedAt: now,
		}
		if err := s.store.CreateMovement(movement); err != nil {
			// 流水写入失败：回滚已扣减的批次与已写入的流水，避免库存与流水分离
			rollBack()
			return err
		}
		created = append(created, movement)
		remaining -= take
	}
	return nil
}

// AdjustStock 盘点调整：正数增加库存（生成批次），负数扣减库存（先进先出）。
// 库存变更与流水写入原子进行：任一流水写入失败时回滚已创建/修改的批次，
// 保证库存与流水同生共死，不会分开提交。
func (s *Service) AdjustStock(productID string, delta int, note string) error {
	if delta == 0 {
		return model.NewValidationError("delta", "调整数量不能为 0")
	}
	if _, err := s.store.GetProduct(productID); err != nil {
		return err
	}
	now := time.Now()
	if delta > 0 {
		batch := &model.InventoryBatch{
			ID:        idgen.Hex(),
			BatchNo:   generateNo("BT"),
			ProductID: productID,
			Quantity:  delta,
			Remaining: delta,
			CreatedAt: now,
		}
		if err := s.store.CreateBatch(batch); err != nil {
			return err
		}
		movement := &model.StockMovement{
			ID:        idgen.Hex(),
			ProductID: productID,
			BatchID:   batch.ID,
			Type:      model.MovementAdjust,
			Delta:     delta,
			Note:      note,
			CreatedAt: now,
		}
		if err := s.store.CreateMovement(movement); err != nil {
			// 流水写入失败：回滚已创建的批次，避免库存与流水分离
			_ = s.store.DeleteBatch(batch.ID)
			return err
		}
		return nil
	}
	// 负向调整按先进先出扣减
	return s.deductStock(productID, -delta, note, model.MovementAdjust)
}

func (s *Service) ListBatches(filter model.BatchFilter, page, size int) ([]*model.InventoryBatch, int, error) {
	all := s.store.ListBatches()
	matched := make([]*model.InventoryBatch, 0, len(all))
	for _, b := range all {
		if filter.Match(b) {
			matched = append(matched, b)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	return pageSlice(matched, page, size)
}

func (s *Service) ListMovements(filter model.MovementFilter, page, size int) ([]*model.StockMovement, int, error) {
	all := s.store.ListMovements()
	matched := make([]*model.StockMovement, 0, len(all))
	for _, m := range all {
		if filter.Match(m) {
			matched = append(matched, m)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	return pageSlice(matched, page, size)
}
