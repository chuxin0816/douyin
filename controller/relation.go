package controller

import (
	"context"
	"douyin/dao/mysql"
	"douyin/pkg/jwt"
	"douyin/response"
	"douyin/service"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type RelationController struct{}

type RelationActionRequest struct {
	Token      string `query:"token" vd:"len($)>0"`                // 用户鉴权token
	ToUserID   int64  `query:"to_user_id,string" vd:"$>0"`         // 对方用户id
	ActionType int    `query:"action_type,string" vd:"$==1||$==2"` // 1-关注，2-取消关注
}

func NewRelationController() *RelationController {
	return &RelationController{}
}

func (rc *RelationController) Action(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &RelationActionRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		hlog.Error("RelationController Action: 参数校验失败, err: ", err)
		response.Error(ctx, response.CodeInvalidParam)
		return
	}

	// 验证token
	userID, err := jwt.ParseToken(req.Token)
	if err != nil {
		response.Error(ctx, response.CodeNoAuthority)
		hlog.Error("RelationController.Action: token无效, err: ", err)
		return
	}

	// 业务逻辑处理
	resp, err := service.RelationAction(userID, req.ToUserID, req.ActionType)
	if err != nil {
		if errors.Is(err, mysql.ErrAlreadyFollow) {
			hlog.Error("RelationController.Action: 已经关注过了, err: ", err)
			response.Error(ctx, response.CodeAlreadyFollow)
			return
		}
		if errors.Is(err, mysql.ErrNotFollow) {
			hlog.Error("RelationController.Action: 还没有关注过, err: ", err)
			response.Error(ctx, response.CodeNotFollow)
			return
		}
		hlog.Error("RelationController.Action: 业务逻辑处理失败, err: ", err)
		response.Error(ctx, response.CodeServerBusy)
		return
	}

	// 返回响应
	response.Success(ctx, resp)
}
