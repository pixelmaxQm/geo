package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

var Platform = new(platform)

type platform struct{}

func (r *platform) Init(public *gin.RouterGroup, private *gin.RouterGroup) {
	group := private.Group("geoMonitor").Use(middleware.OperationRecord())
	{
		group.POST("platform", apiPlatform.CreatePlatform)        // 新增平台
		group.PUT("platform/:id", apiPlatform.UpdatePlatform)     // 编辑平台
		group.DELETE("platform/:id", apiPlatform.DeletePlatform)  // 删除平台
		group.POST("platform/test/:id", apiPlatform.TestPlatform) // 连通性测试
		group.POST("platform/testAll", apiPlatform.TestAllPlatforms) // 一键测试所有
	}
	{
		group := private.Group("geoMonitor")
		group.GET("platform/list", apiPlatform.GetPlatformList) // 平台列表
		group.GET("platform/:id", apiPlatform.GetPlatform)      // 平台详情
	}
}
