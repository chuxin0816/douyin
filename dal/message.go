package dal

import (
	"context"
	"douyin/dal/model"
	"douyin/pkg/snowflake"
	"douyin/rpc/kitex_gen/message"

	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

func MessageAction(ctx context.Context, userID, toUserID int64, content string) error {
	err := qMessage.WithContext(ctx).Create(&model.Message{
		ID:         snowflake.GenerateID(),
		FromUserID: userID,
		ToUserID:   toUserID,
		Content:    content,
		CreateTime: time.Now().Unix(),
	})
	if err != nil {
		klog.Error("插入数据库失败, err: ", err)
		return err
	}

	return nil
}

func MessageList(ctx context.Context, userID, toUserID, lastTime int64) ([]*message.Message, error) {
	mMessageList, err := qMessage.WithContext(ctx).Where(qMessage.FromUserID.Eq(userID), qMessage.ToUserID.Eq(toUserID), qMessage.CreateTime.Gt(lastTime)).
		Or(qMessage.FromUserID.Eq(toUserID), qMessage.ToUserID.Eq(userID), qMessage.CreateTime.Gt(lastTime)).
		Order(qMessage.CreateTime).Find()
	if err != nil {
		klog.Error("查询数据库失败, err: ", err)
		return nil, err
	}

	messageList := make([]*message.Message, len(mMessageList))
	for i, m := range mMessageList {
		messageList[i] = ToMessageResponse(m)
	}

	return messageList, nil
}

func ToMessageResponse(mMessage *model.Message) *message.Message {
	return &message.Message{
		Id:         mMessage.ID,
		ToUserId:   mMessage.ToUserID,
		FromUserId: mMessage.FromUserID,
		Content:    mMessage.Content,
		CreateTime: mMessage.CreateTime,
	}
}
