package initialize

import (
	"context"
	"fmt"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func Gorm(ctx context.Context) {
	// 迁移旧版单列唯一索引为 (code, mode) 复合唯一索引
	if global.GVA_DB.Migrator().HasIndex(&model.Platform{}, "idx_gva_geo_monitor_platforms_code") {
		if err := global.GVA_DB.Migrator().DropIndex(&model.Platform{}, "idx_gva_geo_monitor_platforms_code"); err != nil {
			zap.L().Warn("删除旧唯一索引失败", zap.Error(err))
		}
	}

	err := global.GVA_DB.WithContext(ctx).AutoMigrate(
		new(model.Platform),
		new(model.MonitorTopic),
		new(model.CollectionTask),
		new(model.CollectionResult),
		new(model.CollectionCitation),
		new(model.PlaywrightAuthSession),
	)
	if err != nil {
		err = errors.Wrap(err, "注册表失败!")
		zap.L().Error(fmt.Sprintf("%+v", err))
	}
}
