package main

import (
	"time"

	"github.com/David89Lee/pyrolysis_ty_server/model"
	"go.uber.org/zap"
)

// startHistoryJob 每天凌晨 5 点执行一次历史数据抽稀
func startHistoryJob() {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		lastRunDay := ""
		for now := range ticker.C {
			if now.Hour() != 5 || now.Minute() != 0 {
				continue
			}
			day := now.Format("2006-01-02")
			if day == lastRunDay {
				continue
			}
			lastRunDay = day
			aggregateHistory(now)
		}
	}()
}

// aggregateHistory 把 4 个 DTU 表中一周之前的数据抽稀为 10 分钟一条写入历史表，
// 然后删除实时表中这些数据
func aggregateHistory(now time.Time) {
	cutoff := now.AddDate(0, 0, -7)
	for dtu := uint8(1); dtu <= 4; dtu++ {
		inserted, deleted, err := model.AggregateDtuHistory(dtu, cutoff)
		if err != nil {
			zap.L().Error("历史数据抽稀失败", zap.Uint8("dtu", dtu), zap.Error(err))
			continue
		}
		zap.L().Info("历史数据抽稀完成",
			zap.Uint8("dtu", dtu),
			zap.Int("inserted", inserted),
			zap.Int64("deleted", deleted))
	}
}
