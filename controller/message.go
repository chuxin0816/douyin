package controller

import (
	"context"
	"douyin/middleware"
	"douyin/response"
	"douyin/service"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type MessageController struct{}

type MessageActionRequest struct {
	ToUserID   int64  `query:"to_user_id,string"  vd:"$>0"`      // 对方用户id
	ActionType int64  `query:"action_type,string" vd:"$==1"`     // 1-发送消息
	Content    string `query:"content"            vd:"len($)>0"` // 消息内容
}

type MessageChatRequest struct {
	ToUserID   int64 `query:"to_user_id,string"   vd:"$>0"` // 对方用户id
	PreMsgTime int64 `query:"pre_msg_time,string"`          // 上一条消息时间
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

	// 从认证中间件中获取userID
	userID := ctx.MustGet(middleware.CtxUserIDKey).(int64)

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

	// 从认证中间件中获取userID
	userID := ctx.MustGet(middleware.CtxUserIDKey).(int64)

	// 业务逻辑处理
	resp, err := service.MessageChat(userID, req.ToUserID, req.PreMsgTime)
	if err != nil {
		response.Error(ctx, response.CodeServerBusy)
		hlog.Error("MessageController.Chat: 业务处理失败, err: ", err)
		return
	}

	// 返回响应
	response.Success(ctx, resp)
}
