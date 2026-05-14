package service

var Service = new(service)

type service struct {
	Platform          platform
	Topic             topic
	Collector         collector
	PlaywrightSession playwrightSession
}

