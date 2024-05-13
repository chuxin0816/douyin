package dal

import (
	"context"
	"fmt"
	"sync"
	"time"

	"douyin/src/dal/model"
	"douyin/src/pkg/snowflake"
	"douyin/src/rpc/kitex_gen/message"

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
			Key: "$and", Value: bson.A{
				bson.D{{Key: "convert_id", Value: convertID}},
				bson.D{{Key: "create_time", Value: bson.D{{Key: "$gt", Value: lastTime}}}},
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

	var wgMessageList sync.WaitGroup
	wgMessageList.Add(len(mMessageList))
	messageList := make([]*message.Message, len(mMessageList))
	for i, m := range mMessageList {
		go func(i int, m *model.Message) {
			defer wgMessageList.Done()
			messageList[i] = ToMessageResponse(m)
		}(i, m)
	}
	wgMessageList.Wait()

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