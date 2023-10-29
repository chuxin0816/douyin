package controller

import (
	"context"
	"douyin/pkg/jwt"
	"douyin/response"
	"douyin/service"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type MessageController struct{}

type MessageActionRequest struct {
	Token      string `query:"token"              vd:"len($)>0"` // 用户鉴权token
	ToUserID   int64  `query:"to_user_id,string"  vd:"$>0"`      // 对方用户id
	ActionType int    `query:"action_type,string" vd:"$==1"`     // 1-发送消息
	Content    string `query:"content"            vd:"len($)>0"` // 消息内容
}

type MessageChatRequest struct {
	Token    string `query:"token"             vd:"len($)>0"` // 用户鉴权token
	ToUserID int64  `query:"to_user_id,string" vd:"$>0"`      // 对方用户id
}

func NewMessageController() *MessageController {
	return &MessageController{}
}

func (mc *MessageController) Action(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &MessageActionRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		response.Error(ctx, response.CodeInvalidParam)
		hlog.Error("MessageController.Action: 参数校验失败, err: ", err)
		return
	}

	// 验证token
	userID, err := jwt.ParseToken(req.Token)
	if err != nil {
		response.Error(ctx, response.CodeNoAuthority)
		hlog.Error("MessageController.Action: token无效, err: ", err)
		return
	}

	// 业务逻辑处理
	resp, err := service.MessageAction(userID, req.ToUserID, req.ActionType, req.Content)
	if err != nil {
		response.Error(ctx, response.CodeServerBusy)
		hlog.Error("MessageController.Action: 业务处理失败, err: ", err)
		return
	}

	// 返回响应
	response.Success(ctx, resp)
}

func (mc *MessageController) Chat(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &MessageChatRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		response.Error(ctx, response.CodeInvalidParam)
		hlog.Error("MessageController.Chat: 参数校验失败, err: ", err)
		return
	}

	// 验证token
	userID, err := jwt.ParseToken(req.Token)
	if err != nil {
		response.Error(ctx, response.CodeNoAuthority)
		hlog.Error("MessageController.Chat: token无效, err: ", err)
		return
	}

	// 业务逻辑处理
	resp, err := service.MessageChat(userID, req.ToUserID)
	if err != nil {
		response.Error(ctx, response.CodeServerBusy)
		hlog.Error("MessageController.Chat: 业务处理失败, err: ", err)
		return
	}

	// 返回响应
	response.Success(ctx, resp)
}
