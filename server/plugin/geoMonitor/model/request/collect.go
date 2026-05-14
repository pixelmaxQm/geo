package request

import common "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type RunCollectionRequest struct {
	TopicID     uint   `json:"topicId" binding:"required"`
	PlatformIDs []uint `json:"platformIds" binding:"required,min=1"`
	Mode        string `json:"mode" binding:"required,oneof=api playwright"`
}

type CollectionResultSearch struct {
	common.PageInfo
	TaskID uint `json:"taskId" form:"taskId" binding:"required"`
}
