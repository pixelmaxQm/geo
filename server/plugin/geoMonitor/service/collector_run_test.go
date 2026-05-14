package service

import (
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model"
	geoReq "github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model/request"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCollectorRunRejectsForeignTopicForNonSuperAdmin(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	oldDB := global.GVA_DB
	global.GVA_DB = db
	defer func() { global.GVA_DB = oldDB }()

	if err := db.AutoMigrate(&model.MonitorTopic{}, &model.Platform{}, &model.CollectionTask{}, &model.CollectionResult{}); err != nil {
		t.Fatalf("migrate schema: %v", err)
	}

	topic := model.MonitorTopic{Type: "fact_check", Name: "topic-a", Prompt: "prompt-a", Status: 1, UserID: 99, UserName: "owner"}
	if err := db.Create(&topic).Error; err != nil {
		t.Fatalf("create topic: %v", err)
	}

	collectorService := collector{}
	_, err = collectorService.Run(geoReq.RunCollectionRequest{
		TopicID:     topic.ID,
		PlatformIDs: []uint{1},
		Mode:        CollectModeAPI,
	}, 1, "tester", 100)
	if err == nil {
		t.Fatal("expected permission-scoped topic lookup to fail")
	}
}
