package mysql

import (
	"douyin/models"
	"douyin/pkg/snowflake"
	"douyin/response"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func MessageAction(userID, toUserID int64, content string) error {
	err := db.Create(&models.Message{
		ID:         snowflake.GenerateID(),
		FromUserID: userID,
		ToUserID:   toUserID,
		Content:    content,
		CreateTime: time.Now(),
	}).Error

	return err
}

func MessageList(userID, toUserID int64) ([]*response.MessageResponse, error) {
	var dMessageList []*models.Message
	err := db.Where("from_user_id = ? and to_user_id = ?", userID, toUserID).Or("from_user_id = ? and to_user_id = ?", toUserID, userID).Order("").Find(&dMessageList).Error
	if err != nil {
		hlog.Error("mysql.MessageList: 查询数据库失败, err: ", err)
		return nil, err
	}

	messageList := make([]*response.MessageResponse, 0, len(dMessageList))
	for _, message := range dMessageList {
		messageList = append(messageList, ToMessageResponse(message))
	}

	return messageList, nil
}

func ToMessageResponse(message *models.Message) *response.MessageResponse {
	return &response.MessageResponse{
		ID:         message.ID,
		ToUserID:   message.ToUserID,
		FromUserID: message.FromUserID,
		Content:    message.Content,
		CreateTime: message.CreateTime.Unix(),
	}
}
