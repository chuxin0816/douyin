package main

import (
	"context"
	"douyin/dal"
	message "douyin/rpc/kitex_gen/message"

	"github.com/u2takey/go-utils/klog"
)

// MessageServiceImpl implements the last service interface defined in the IDL.
type MessageServiceImpl struct{}

// MessageChat implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) MessageChat(ctx context.Context, req *message.MessageChatRequest) (resp *message.MessageChatResponse, err error) {
	// 操作数据库
	messageList, err := dal.MessageList(ctx, req.UserId, req.ToUserId, req.LastTime)
	if err != nil {
		klog.Error("操作数据库失败, err: ", err)
		return nil, err
	}

	// 返回响应
	resp = &message.MessageChatResponse{MessageList: messageList}

	return
}

// MessageAction implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) MessageAction(ctx context.Context, req *message.MessageActionRequest) (resp *message.MessageActionResponse, err error) {
	// 操作数据库
	if err := dal.MessageAction(ctx, req.UserId, req.ToUserId, req.Content); err != nil {
		klog.Error("操作数据库失败, err: ", err)
		return nil, err
	}

	// 返回响应
	resp = &message.MessageActionResponse{}

	return
}
