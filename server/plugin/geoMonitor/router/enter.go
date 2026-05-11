package router

import "github.com/flipped-aurora/gin-vue-admin/server/plugin/geoMonitor/api"

var (
	Router      = new(router)
	apiPlatform = api.Api.Platform
)

type router struct{ Platform platform }
