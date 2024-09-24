package dal

import (
	"context"
	"fmt"
	"time"

	"douyin/src/common/snowflake"
	"douyin/src/dal/model"

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

func MessageList(ctx context.Context, userID, toUserID, lastTime int64) ([]*model.Message, error) {
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

	messageList := make([]*model.Message, 0)
	for cursor.Next(ctx) {
		var m model.Message
		if err := cursor.Decode(&m); err != nil {
			return nil, err
		}
		messageList = append(messageList, &m)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return messageList, nil
}
