package router

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/api"

var (
	Router               = new(router)
	apiPlatform          = api.Api.Platform
	apiTopic             = api.Api.Topic
	apiCollect           = api.Api.Collect
	apiPlaywrightSession = api.Api.PlaywrightSession
)

type router struct {
	Platform          platform
	Topic             topic
	Collect           collect
	PlaywrightSession playwrightSession
}
