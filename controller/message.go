package controller

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

type MessageController struct{}

type MessageActionRequest struct {
	Token      string `query:"token"              vd:"len($)>0"` // 用户鉴权token
	ToUserID   int64  `query:"to_user_id,string"  vd:"$>0"`      // 对方用户id
	ActionType int    `query:"action_type,string" vd:"$==1"`     // 1-发送消息
	Content    string `query:"content"            vd:"len($)>0"` // 消息内容
}

func NewMessageController() *MessageController {
	return &MessageController{}
}

func (mc *MessageController) Action(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &MessageActionRequest{}
	ctx.BindAndValidate(req)
}
