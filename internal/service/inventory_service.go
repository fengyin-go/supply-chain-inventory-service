package service

import (
	"errors"
	"sort"
	"time"

	"supplychain/internal/model"
	"supplychain/internal/store"
	"supplychain/pkg/idgen"
)

// StockInbound 入库：质检合格后，将质检中的入库单转为已入库，
// 并生成库存批次与入库流水，同时将采购单标记为已收货。
//
// 并发安全：同一张入库单的两次记账请求会同时读到「质检中」状态。为避免两边
// 各自生成批次与流水导致重复入库，这里用 TransitionInboundStatus 在 store 层
// 原子地完成「状态仍是质检中 → 置为已入库」的占位：只有先到的那次能成功，后到
// 的一次因状态已被改掉拿到 ErrConflict 而直接返回，不会进入批次/流水创建逻辑。
func (s *Service) StockInbound(inboundID string) (*model.InboundOrder, error) {
	// 仅校验入库单是否存在：缺失时由 store 返回 ErrNotFound，交由上层映射为 404。
	// 这里不读取/依赖返回的入库单状态做判定，避免与并发的状态写入产生数据竞争；
	// 状态判定统一交给下面的原子占位完成。
	if _, err := s.store.GetInboundOrder(inboundID); err != nil {
		return nil, err
	}
	ins, err := s.store.GetInspectionByInbound(inboundID)
	if err != nil {
		return nil, model.NewValidationError("inspection", "入库前必须先完成质检")
	}
	if ins.Result != model.InspectionPassed {
		return nil, model.NewValidationError("inspection", "质检不合格，不能入库")
	}
	// 原子占位：质检中 → 已入库。返回的 inbound 已是占位成功后的最新值，
	// 后续批次/流水创建只读其 Items 等创建后不变的字段，不再触碰 Status。
	inbound, err := s.store.TransitionInboundStatus(inboundID, model.InboundInspecting, model.InboundStocked)
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			return nil, model.NewValidationError("status", "该入库单已入库，请勿重复入库")
		}
		return nil, err
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
func (s *Service) OutboundStock(productID string, quantity int, note string) error {
	if quantity <= 0 {
		return model.NewValidationError("quantity", "出库数量必须大于 0")
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
	for _, b := range batches {
		if remaining == 0 {
			break
		}
		if b.Remaining == 0 {
			continue
		}
		take := b.Remaining
		if take > remaining {
			take = remaining
		}
		b.Remaining -= take
		if err := s.store.UpdateBatch(b); err != nil {
			return err
		}
		movement := &model.StockMovement{
			ID:        idgen.Hex(),
			ProductID: productID,
			BatchID:   b.ID,
			Type:      model.MovementOutbound,
			Delta:     -take,
			Note:      note,
			CreatedAt: now,
		}
		if err := s.store.CreateMovement(movement); err != nil {
			return err
		}
		remaining -= take
	}
	return nil
}

// AdjustStock 盘点调整：正数增加库存（生成批次），负数扣减库存（先进先出）。
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
		return s.store.CreateMovement(&model.StockMovement{
			ID:        idgen.Hex(),
			ProductID: productID,
			BatchID:   batch.ID,
			Type:      model.MovementAdjust,
			Delta:     delta,
			Note:      note,
			CreatedAt: now,
		})
	}
	// 负向调整按先进先出扣减
	batches := s.store.ListBatchesByProduct(productID)
	sort.Slice(batches, func(i, j int) bool {
		return batches[i].CreatedAt.Before(batches[j].CreatedAt)
	})
	need := -delta
	available := 0
	for _, b := range batches {
		available += b.Remaining
	}
	if available < need {
		return model.NewValidationError("delta", "库存不足，无法扣减")
	}
	remaining := need
	for _, b := range batches {
		if remaining == 0 {
			break
		}
		if b.Remaining == 0 {
			continue
		}
		take := b.Remaining
		if take > remaining {
			take = remaining
		}
		b.Remaining -= take
		if err := s.store.UpdateBatch(b); err != nil {
			return err
		}
		if err := s.store.CreateMovement(&model.StockMovement{
			ID:        idgen.Hex(),
			ProductID: productID,
			BatchID:   b.ID,
			Type:      model.MovementAdjust,
			Delta:     -take,
			Note:      note,
			CreatedAt: now,
		}); err != nil {
			return err
		}
		remaining -= take
	}
	return nil
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
