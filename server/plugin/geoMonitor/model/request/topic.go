package request

import "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type TopicSearch struct {
	request.PageInfo
	Type   string `json:"type" form:"type"`
	Name   string `json:"name" form:"name"`
	Status *int   `json:"status" form:"status"`
}
