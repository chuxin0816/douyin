package mysql

import (
	"douyin/models"
	"time"
)

func MessageAction(userID, toUserID int64, content string) error {
	err := db.Create(&models.Message{
		FromUserID: userID,
		ToUserID:   toUserID,
		Content:    content,
		CreateTime: time.Now(),
	}).Error
	
	return err
}
