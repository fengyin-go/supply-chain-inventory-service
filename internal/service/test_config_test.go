package service

import (
	"supplychain/internal/config"
	"supplychain/pkg/logger"
)

func testLogger() *logger.Logger { return logger.NewLevel(logger.LevelError) }
func testConfig() *config.Config { return &config.Config{MaxPageSize: 100} }
