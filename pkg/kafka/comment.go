package kafka

import (
	"context"
	"douyin/config"
	"douyin/dal"
	"douyin/dal/model"
	"encoding/json"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type commentMQ struct {
	*mq
}

var commentMQInstance *commentMQ

func initCommentMQ() {
	commentMQInstance = &commentMQ{
		&mq{
			Topic:  topicComment,
			Writer: NewWriter(topicComment),
			Reader: NewReader(topicComment),
		},
	}

	go commentMQInstance.consumeComment(context.Background())
}

func (mq *commentMQ) consumeComment(ctx context.Context) {
	_, span := otel.Tracer(config.Conf.OpenTelemetryConfig.KafkaName).Start(ctx, "kafka.consumeComment")
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

		// 解析为Comment，成功则创建评论
		comment := &model.Comment{}
		if err := json.Unmarshal(m.Value, comment); err == nil {
			if err := dal.CreateComment(ctx, comment); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "failed to create comment")
				klog.Error("failed to create comment: ", err)
			}
			continue
		}

		// 解析为CommentID，成功则删除评论
		var commentID int64
		if err := json.Unmarshal(m.Value, &commentID); err == nil {
			if err := dal.DeleteComment(ctx, commentID); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "failed to delete comment")
				klog.Error("failed to delete comment: ", err)
			}
			continue
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to unmarshal message")
		klog.Error("failed to unmarshal message: ", err)
	}
	// 程序退出前关闭Reader
	if err := mq.Reader.Close(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to close reader")
		klog.Fatal("failed to close reader:", err)
	}
}

func CreateComment(ctx context.Context, comment *model.Comment) error {
	_, span := otel.Tracer(config.Conf.OpenTelemetryConfig.KafkaName).Start(ctx, "kafka.CreateComment")
	defer span.End()

	value, err := json.Marshal(comment)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal message")
		klog.Error("failed to marshal message: ", err)
		return err
	}
	return commentMQInstance.Writer.WriteMessages(ctx, kafka.Message{
		Value: value,
	})
}

func DeleteComment(ctx context.Context, commentID int64) error {
	_, span := otel.Tracer(config.Conf.OpenTelemetryConfig.KafkaName).Start(ctx, "kafka.DeleteComment")
	defer span.End()

	value, err := json.Marshal(commentID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal message")
		klog.Error("failed to marshal message: ", err)
		return err
	}
	return commentMQInstance.Writer.WriteMessages(ctx, kafka.Message{
		Value: value,
	})
}
