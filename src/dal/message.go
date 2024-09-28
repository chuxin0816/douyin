package dal

import (
	"context"
	"strconv"
	"strings"
	"time"

	"douyin/src/common/snowflake"
	"douyin/src/dal/model"
)

func MessageAction(ctx context.Context, message *model.Message) error {
	message.ID = snowflake.GenerateID()
	message.CreateTime = time.Now().UnixMilli()

	return db.WithContext(ctx).Model(&model.Message{}).Create(message).Error
}

func MessageList(ctx context.Context, userID, toUserID, lastTime int64) ([]*model.Message, error) {
	convertID := GetConvertID(userID, toUserID)

	messageList := make([]*model.Message, 0)
	err := db.WithContext(ctx).Model(&model.Message{}).Where("convert_id = ? AND create_time > ?", convertID, lastTime).Find(&messageList).Error

	return messageList, err
}

func GetConvertID(userID, toUserID int64) string {
	var builder strings.Builder
	if userID < toUserID {
		builder.WriteString(strconv.FormatInt(userID, 10))
		builder.WriteString("_")
		builder.WriteString(strconv.FormatInt(toUserID, 10))
	} else {
		builder.WriteString(strconv.FormatInt(toUserID, 10))
		builder.WriteString("_")
		builder.WriteString(strconv.FormatInt(userID, 10))
	}
	return builder.String()
}
