package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

var Topic = new(topic)

type topic struct{}

func (r *topic) Init(public *gin.RouterGroup, private *gin.RouterGroup) {
	{
		group := private.Group("geoMonitor").Use(middleware.OperationRecord())
		group.POST("topic", apiTopic.CreateTopic)       // 新增话题
		group.PUT("topic/:id", apiTopic.UpdateTopic)     // 编辑话题
		group.DELETE("topic/:id", apiTopic.DeleteTopic)  // 删除话题
	}
	{
		group := private.Group("geoMonitor")
		group.GET("topic/list", apiTopic.GetTopicList)  // 话题列表
		group.GET("topic/:id", apiTopic.GetTopic)       // 话题详情
	}
}
