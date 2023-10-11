package controller

import (
	"context"

	"github.com/chuxin0816/Scaffold/models"
	"github.com/chuxin0816/Scaffold/service"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// Feed 不限制登录状态，返回按投稿时间倒序的视频列表，视频数由服务端控制，单次最多30个
func Feed(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &models.FeedRequest{}
	ctx.Bind(req)

	//TODO: 业务逻辑
	resp, err := service.Feed(req)
	if err != nil {
		ctx.JSON(consts.StatusInternalServerError, nil)
		hlog.Errorf("Feed service error: %v", err)
		return
	}

	// 返回结果
	ctx.JSON(consts.StatusOK, resp)
}
