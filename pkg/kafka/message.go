package kafka

import (
	"context"
	"encoding/json"

	"douyin/dal"
	"douyin/dal/model"
	"douyin/pkg/tracing"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/codes"
)

type messageMQ struct {
	*mq
}

var messageMQInstance *messageMQ

func initMessageMQ() {
	messageMQInstance = &messageMQ{
		&mq{
			Topic:  topicMessage,
			Writer: NewWriter(topicMessage),
			Reader: NewReader(topicMessage),
		},
	}

	go messageMQInstance.consumeMessage(context.Background())
}

func (mq *messageMQ) consumeMessage(ctx context.Context) {
	// 接收消息
	for {
		m, err := mq.Reader.ReadMessage(ctx)
		if err != nil {
			klog.Error("failed to read message: ", err)
			break
		}

		message := &model.Message{}
		if err := json.Unmarshal(m.Value, message); err != nil {
			klog.Error("failed to unmarshal message: ", err)
			continue
		}

		// 写入数据库
		if err := dal.MessageAction(ctx, message); err != nil {
			klog.Error("failed to write message to database: ", err)
		}
	}
}

func SendMessage(ctx context.Context, message *model.Message) error {
	ctx, span := tracing.Tracer.Start(ctx, "kafka.SendMessage")
	defer span.End()

	data, err := json.Marshal(message)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal message")
		klog.Error("failed to marshal message: ", err)
		return err
	}

	return favoriteMQInstance.Writer.WriteMessages(ctx, kafka.Message{
		Value: data,
	})
}
