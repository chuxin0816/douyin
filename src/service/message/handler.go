package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"douyin/src/dal"
	"douyin/src/dal/model"
	message "douyin/src/kitex_gen/message"
	"douyin/src/pkg/kafka"
	"douyin/src/pkg/tracing"

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
	mMessageList, err := dal.MessageList(ctx, req.UserId, req.ToUserId, req.LastTime)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "操作数据库失败")
		klog.Error("操作数据库失败, err: ", err)
		return nil, err
	}

	// 将model.Message转换为message.Message
	var wg sync.WaitGroup
	wg.Add(len(mMessageList))
	messageList := make([]*message.Message, len(mMessageList))
	for i, m := range mMessageList {
		go func(i int, m *model.Message) {
			defer wg.Done()
			messageList[i] = toMessageResponse(m)
		}(i, m)
	}
	wg.Wait()

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

	msg := &model.Message{
		ToUserID:   req.ToUserId,
		FromUserID: req.UserId,
		ConvertID:  convertID,
		Content:    req.Content,
		CreateTime: time.Now().Unix(),
	}
	// 通过kafka更新数据库
	if err := kafka.SendMessage(ctx, msg); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "通过kafka更新数据库失败")
		klog.Error("通过kafka更新数据库失败, err: ", err)
		return nil, err
	}

	// 返回响应
	resp = &message.MessageActionResponse{}

	return
}

func toMessageResponse(mMessage *model.Message) *message.Message {
	return &message.Message{
		Id:         mMessage.ID,
		ToUserId:   mMessage.ToUserID,
		FromUserId: mMessage.FromUserID,
		Content:    mMessage.Content,
		CreateTime: mMessage.CreateTime,
	}
}
