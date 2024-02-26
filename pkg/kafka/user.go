package kafka

import (
	"context"
	"douyin/dal"
	"douyin/dal/model"
	"encoding/json"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/segmentio/kafka-go"
)

type userMQ struct {
	*mq
}

var userMQInstance *userMQ

func initUserMQ() {
	userMQInstance = &userMQ{
		mq: &mq{
			Topic:  topicUser,
			Writer: NewWriter(topicUser),
			Reader: NewReader(topicUser),
		},
	}
}

func (mq *userMQ) consumeUser(ctx context.Context) {
	for {
		m, err := mq.Reader.ReadMessage(ctx)
		if err != nil {
			klog.Error("failed to read message: ", err)
			break
		}
		user := &model.User{}
		if err := json.Unmarshal(m.Value, user); err != nil {
			klog.Error("failed to unmarshal message: ", err)
			continue
		}
		// 更新数据
		if err := dal.UpdateUser(ctx, user); err != nil {
			klog.Error("failed to update user: ", err)
			continue
		}
	}
	// 程序退出前关闭Reader
	if err := mq.Reader.Close(); err != nil {
		klog.Fatal("failed to close reader:", err)
	}
}

func UpdateUser(user *model.User) error {
	value, err := json.Marshal(user)
	if err != nil {
		klog.Error("failed to marshal message:", err)
		return err
	}

	return userMQInstance.Writer.WriteMessages(context.Background(), kafka.Message{
		Value: value,
	})
}
