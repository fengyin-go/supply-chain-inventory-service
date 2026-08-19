// Package service 实现业务逻辑层。
package service

import (
	"supplychain/internal/config"
	"supplychain/internal/store"
	"supplychain/pkg/logger"
)

type Service struct {
	store store.Store
	log   *logger.Logger
	cfg   *config.Config
}

func New(st store.Store, log *logger.Logger, cfg *config.Config) *Service {
	return &Service{store: st, log: log, cfg: cfg}
}
