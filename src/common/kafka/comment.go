package kafka

import (
	"context"

	"douyin/src/dal"
	"douyin/src/dal/model"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/segmentio/kafka-go"
	"github.com/vmihailenco/msgpack/v5"
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
	// 接收消息
	for {
		ctx, span := otel.Tracer("kafka").Start(ctx, "consumeComment")

		m, err := mq.Reader.FetchMessage(ctx)
		if err != nil {
			klog.Error("failed to fetch message: ", err)
			span.End()
			break
		}

		// 解析为Comment，成功则创建评论
		comment := &model.Comment{}
		if err := msgpack.Unmarshal(m.Value, comment); err == nil {
			if err := dal.CreateComment(ctx, comment); err != nil {
				klog.Error("failed to create comment: ", err)
				span.End()
				continue
			}
		}

		// 解析为CommentID，成功则删除评论
		var commentID int64
		if err := msgpack.Unmarshal(m.Value, &commentID); err == nil {
			if err := dal.DeleteComment(ctx, commentID); err != nil {
				klog.Error("failed to delete comment: ", err)
				span.End()
				continue
			}
		} else {
			klog.Error("failed to unmarshal message: ", err)
			span.End()
			continue
		}

		if err := mq.Reader.CommitMessages(ctx, m); err != nil {
			klog.Error("failed to commit message: ", err)
		}

		span.End()
	}

	// 程序退出前关闭Reader
	if err := mq.Reader.Close(); err != nil {
		klog.Fatal("failed to close reader:", err)
	}
}

func CreateComment(ctx context.Context, comment *model.Comment) error {
	ctx, span := otel.Tracer("kafka").Start(ctx, "CreateComment")
	defer span.End()

	value, err := msgpack.Marshal(comment)
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
	ctx, span := otel.Tracer("kafka").Start(ctx, "DeleteComment")
	defer span.End()

	value, err := msgpack.Marshal(commentID)
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
