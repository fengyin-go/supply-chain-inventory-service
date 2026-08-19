package service

import (
	"sort"
	"time"

	"supplychain/internal/model"
)

func (s *Service) GetInspection(id string) (*model.Inspection, error) {
	return s.store.GetInspection(id)
}

func (s *Service) ListInspections(filter model.InspectionFilter, page, size int) ([]*model.Inspection, int, error) {
	all := s.store.ListInspections()
	matched := make([]*model.Inspection, 0, len(all))
	for _, i := range all {
		if filter.Match(i) {
			matched = append(matched, i)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	return pageSlice(matched, page, size)
}

// CompleteInspection 完成质检：待质检 → 合格/不合格。
func (s *Service) CompleteInspection(id, result string, defectCount int, note string) (*model.Inspection, error) {
	ins, err := s.store.GetInspection(id)
	if err != nil {
		return nil, err
	}
	if !model.CanTransitionInspection(ins.Result, result) {
		return nil, model.NewValidationError("result", "当前状态不可流转")
	}
	if defectCount < 0 {
		return nil, model.NewValidationError("defect_count", "缺陷数量不能为负数")
	}
	ins.Result = result
	ins.DefectCount = defectCount
	ins.Note = note
	ins.UpdatedAt = time.Now()
	if err := s.store.UpdateInspection(ins); err != nil {
		return nil, err
	}
	return ins, nil
}

func (s *Service) DeleteInspection(id string) error {
	return s.store.DeleteInspection(id)
}
