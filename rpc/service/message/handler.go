package main

import (
	"context"
	"douyin/config"
	"douyin/dal"
	message "douyin/rpc/kitex_gen/message"

	"github.com/cloudwego/kitex/pkg/klog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

// MessageServiceImpl implements the last service interface defined in the IDL.
type MessageServiceImpl struct{}

// MessageChat implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) MessageChat(ctx context.Context, req *message.MessageChatRequest) (resp *message.MessageChatResponse, err error) {
	ctx, span := otel.Tracer(config.Conf.OpenTelemetryConfig.MessageName).Start(ctx, "rpc.MessageChat")
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
	ctx, span := otel.Tracer(config.Conf.OpenTelemetryConfig.MessageName).Start(ctx, "rpc.MessageAction")
	defer span.End()

	// 操作数据库
	if err := dal.MessageAction(ctx, req.UserId, req.ToUserId, req.Content); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "操作数据库失败")
		klog.Error("操作数据库失败, err: ", err)
		return nil, err
	}

	// 返回响应
	resp = &message.MessageActionResponse{}

	return
}
