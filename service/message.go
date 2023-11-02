package service

import (
	"douyin/dao/mysql"
	"douyin/response"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func MessageAction(userID, toUserID int64, actionType int64, content string) (*response.MessageActionResponse, error) {
	// 操作数据库
	err := mysql.MessageAction(userID, toUserID, content)
	if err != nil {
		hlog.Error("service.MessageAction: 操作数据库失败, err: ", err)
		return nil, err
	}

	// 返回响应
	return &response.MessageActionResponse{
		Response: &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
	}, nil
}

func MessageChat(userID, toUserID, lastTime int64) (*response.MessageChatResponse, error) {
	// 操作数据库
	messageList, err := mysql.MessageList(userID, toUserID, lastTime)
	if err != nil {
		hlog.Error("service.MessageChat: 操作数据库失败, err: ", err)
		return nil, err
	}

	// 返回响应
	return &response.MessageChatResponse{
		Response:    &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
		MessageList: messageList,
	}, nil
}
