package api

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/service"

var (
	Api             = new(api)
	platformService = service.Service.Platform
)

type api struct{ Platform platform }
