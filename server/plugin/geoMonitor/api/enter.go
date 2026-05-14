package api

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/service"

var (
	Api                      = new(api)
	platformService          = service.Service.Platform
	topicService             = service.Service.Topic
	collectorService         = service.Service.Collector
	playwrightSessionService = service.Service.PlaywrightSession
)

type api struct {
	Platform          platform
	Topic             topic
	Collect           collect
	PlaywrightSession playwrightSession
}
