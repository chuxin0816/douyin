package controller

import (
	"context"

	"douyin/src/client"

	"douyin/src/kitex_gen/message"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
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
	c, span := otel.Tracer("message").Start(c, "MessageAction")
	defer span.End()

	// 获取参数
	req := &MessageActionRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		span.RecordError(err)
		span.SetStatus(codes.Error, "参数校验失败")
		hlog.Error("参数校验失败, err: ", err)
		return
	}

	// 从认证中间件中获取userID
	userID := ctx.MustGet(CtxUserIDKey).(int64)

	// 业务逻辑处理
	resp, err := client.MessageClient.MessageAction(c, &message.MessageActionRequest{
		UserId:     userID,
		ToUserId:   req.ToUserID,
		ActionType: req.ActionType,
		Content:    req.Content,
	})
	if err != nil {
		Error(ctx, CodeServerBusy)
		span.RecordError(err)
		span.SetStatus(codes.Error, "业务处理失败")
		hlog.Error("业务处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}

func (mc *MessageController) Chat(c context.Context, ctx *app.RequestContext) {
	c, span := otel.Tracer("message").Start(c, "MessageChat")
	defer span.End()

	// 获取参数
	req := &MessageChatRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		span.RecordError(err)
		span.SetStatus(codes.Error, "参数校验失败")
		hlog.Error("参数校验失败, err: ", err)
		return
	}

	// 从认证中间件中获取userID
	userID := ctx.MustGet(CtxUserIDKey).(int64)

	// 业务逻辑处理
	resp, err := client.MessageClient.MessageChat(c, &message.MessageChatRequest{
		UserId:   userID,
		ToUserId: req.ToUserID,
		LastTime: req.PreMsgTime,
	})
	if err != nil {
		Error(ctx, CodeServerBusy)
		span.RecordError(err)
		span.SetStatus(codes.Error, "业务处理失败")
		hlog.Error("业务处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}
