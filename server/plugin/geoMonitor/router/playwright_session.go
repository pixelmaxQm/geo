package router

import "github.com/gin-gonic/gin"

type playwrightSession struct{}

func (r *playwrightSession) Init(public *gin.RouterGroup, private *gin.RouterGroup) {
	group := private.Group("geoMonitor")
	group.POST("playwright/session/start", apiPlaywrightSession.Start)
	group.GET("playwright/session/:id", apiPlaywrightSession.Get)
	group.POST("playwright/session/:id/refresh", apiPlaywrightSession.Refresh)
	group.DELETE("playwright/session/:id", apiPlaywrightSession.Delete)
}
