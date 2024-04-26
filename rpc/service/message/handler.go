package main

import (
	"context"
	"fmt"
	"time"

	"douyin/dal"
	"douyin/dal/model"
	"douyin/pkg/tracing"
	message "douyin/rpc/kitex_gen/message"

	"github.com/cloudwego/kitex/pkg/klog"
	"go.opentelemetry.io/otel/codes"
)

// MessageServiceImpl implements the last service interface defined in the IDL.
type MessageServiceImpl struct{}

// MessageChat implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) MessageChat(ctx context.Context, req *message.MessageChatRequest) (resp *message.MessageChatResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "MessageChat")
	defer span.End()

	// 操作数据库
	messageList, err := dal.MessageList(ctx, req.UserId, req.ToUserId, req.LastTime)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "操作数据库失败")
		klog.Error("操作数据库失败, err: ", err)
		return nil, err
	}

	// 返回响应
	resp = &message.MessageChatResponse{MessageList: messageList}

	return
}

// MessageAction implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) MessageAction(ctx context.Context, req *message.MessageActionRequest) (resp *message.MessageActionResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "MessageAction")
	defer span.End()

	var convertID string
	if req.UserId < req.ToUserId {
		convertID = fmt.Sprintf("%d_%d", req.UserId, req.ToUserId)
	} else {
		convertID = fmt.Sprintf("%d_%d", req.ToUserId, req.UserId)
	}

	// 操作数据库
	msg := &model.Message{
		ToUserID:   req.ToUserId,
		FromUserID: req.UserId,
		ConvertID:  convertID,
		Content:    req.Content,
		CreateTime: time.Now().Unix(),
	}
	if err := dal.MessageAction(ctx, msg); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "操作数据库失败")
		klog.Error("操作数据库失败, err: ", err)
		return nil, err
	}

	// 返回响应
	resp = &message.MessageActionResponse{}

	return
}
