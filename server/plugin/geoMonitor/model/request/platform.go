package request

import "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type PlatformSearch struct {
	request.PageInfo
	Code   string `json:"code" form:"code"`
	Name   string `json:"name" form:"name"`
	Status *int   `json:"status" form:"status"`
}
