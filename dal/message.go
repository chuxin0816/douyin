package dal

import (
	"context"
	"fmt"
	"time"

	"douyin/dal/model"
	"douyin/pkg/snowflake"
	"douyin/rpc/kitex_gen/message"

	"go.mongodb.org/mongo-driver/bson"
)



func MessageAction(ctx context.Context, message *model.Message) error {
	message.ID = snowflake.GenerateID()
	message.CreateTime = time.Now().Unix()

	_, err := collectionMessage.InsertOne(ctx, message)
	if err != nil {
		return err
	}

	return nil
}

func MessageList(ctx context.Context, userID, toUserID, lastTime int64) ([]*message.Message, error) {
	var convertID string
	if userID < toUserID {
		convertID = fmt.Sprintf("%d_%d", userID, toUserID)
	} else {
		convertID = fmt.Sprintf("%d_%d", toUserID, userID)
	}
	// 查询条件
	filter := bson.D{
		{
			"$and", bson.A{
				bson.D{{"convert_id", convertID}},
				bson.D{{"create_time", bson.D{{"$gt", lastTime}}}},
			},
		},
	}

	cursor, err := collectionMessage.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	mMessageList := make([]*model.Message, 0)
	for cursor.Next(ctx) {
		var m model.Message
		if err := cursor.Decode(&m); err != nil {
			return nil, err
		}
		mMessageList = append(mMessageList, &m)
	}
	if err := cursor.Err(); err != nil {
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
