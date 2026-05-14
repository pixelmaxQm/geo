package router

import "github.com/gin-gonic/gin"

type collect struct{}

func (r *collect) Init(public *gin.RouterGroup, private *gin.RouterGroup) {
	group := private.Group("geoMonitor")
	group.POST("collect/run", apiCollect.Run)
	group.GET("collect/task/:id", apiCollect.GetTask)
	group.GET("collect/result/list", apiCollect.GetResultList)
}
