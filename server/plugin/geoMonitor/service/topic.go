package service

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/model/request"
)

var Topic = new(topic)

type topic struct{}

const superAdminAuthorityID uint = 888

func (s *topic) GetTopicList(info request.TopicSearch, userID uint, authorityID uint) (list []model.MonitorTopic, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)

	db := global.GVA_DB.Model(&model.MonitorTopic{})
	if authorityID != superAdminAuthorityID {
		db = db.Where("user_id = ?", userID)
	}
	if info.Type != "" {
		db = db.Where("type = ?", info.Type)
	}
	if info.Name != "" {
		db = db.Where("name LIKE ?", "%"+info.Name+"%")
	}
	if info.Status != nil {
		db = db.Where("status = ?", *info.Status)
	}

	err = db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.Limit(limit).Offset(offset).Order("id desc").Find(&list).Error
	return list, total, err
}

func (s *topic) GetTopic(id uint, userID uint, authorityID uint) (info model.MonitorTopic, err error) {
	db := global.GVA_DB.Where("id = ?", id)
	if authorityID != superAdminAuthorityID {
		db = db.Where("user_id = ?", userID)
	}
	err = db.First(&info).Error
	return
}

func (s *topic) CreateTopic(info *model.MonitorTopic) error {
	return global.GVA_DB.Create(info).Error
}

func (s *topic) UpdateTopic(info *model.MonitorTopic, userID uint, authorityID uint) error {
	db := global.GVA_DB.Model(&model.MonitorTopic{}).Where("id = ?", info.ID)
	if authorityID != superAdminAuthorityID {
		db = db.Where("user_id = ?", userID)
	}
	return db.Updates(info).Error
}

func (s *topic) DeleteTopic(id uint, userID uint, authorityID uint) error {
	db := global.GVA_DB.Where("id = ?", id)
	if authorityID != superAdminAuthorityID {
		db = db.Where("user_id = ?", userID)
	}
	return db.Delete(&model.MonitorTopic{}).Error
}
