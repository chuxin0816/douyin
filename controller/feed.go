package controller

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

// Feed 不限制登录状态，返回按投稿时间倒序的视频列表，视频数由服务端控制，单次最多30个
func Feed(c context.Context, ctx *app.RequestContext) {
	// 获取参数

}
