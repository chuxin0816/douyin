package kafka

import (
	"context"
	"douyin/dal"
	"douyin/dal/model"
	"douyin/pkg/tracing"
	"encoding/json"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/codes"
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

	go userMQInstance.consumeUser(context.Background())
}

func (mq *userMQ) consumeUser(ctx context.Context) {
	_, span := tracing.Tracer.Start(ctx, "kafka.consumeUser")
	defer span.End()

	// 接收消息
	for {
		m, err := mq.Reader.ReadMessage(ctx)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to read message")
			klog.Error("failed to read message: ", err)
			break
		}
		user := &model.User{}
		if err := json.Unmarshal(m.Value, user); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to unmarshal message")
			klog.Error("failed to unmarshal message: ", err)
			continue
		}
		// 更新数据
		if err := dal.UpdateUser(ctx, user); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to update user")
			klog.Error("failed to update user: ", err)
			continue
		}
	}
	// 程序退出前关闭Reader
	if err := mq.Reader.Close(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to close reader")
		klog.Fatal("failed to close reader:", err)
	}
}

func UpdateUser(user *model.User) error {
	_, span := tracing.Tracer.Start(context.Background(), "kafka.UpdateUser")
	defer span.End()

	value, err := json.Marshal(user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal message")
		klog.Error("failed to marshal message:", err)
		return err
	}

	return userMQInstance.Writer.WriteMessages(context.Background(), kafka.Message{
		Value: value,
	})
}
