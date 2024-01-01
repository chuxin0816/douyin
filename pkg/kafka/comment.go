package kafka

import (
	"context"
	"douyin/dal"
	"douyin/dal/model"
	"encoding/json"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/segmentio/kafka-go"
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
		m, err := mq.Reader.ReadMessage(ctx)
		if err != nil {
			klog.Error("failed to read message: ", err)
			break
		}

		// 解析为Comment，成功则创建评论
		comment := &model.Comment{}
		if err := json.Unmarshal(m.Value, comment); err == nil {
			if err := dal.CreateComment(ctx, comment); err != nil {
				klog.Error("failed to create comment: ", err)
			}
			continue
		}

		// 解析为CommentID，成功则删除评论
		var commentID int64
		if err := json.Unmarshal(m.Value, commentID); err == nil {
			if err := dal.DeleteComment(ctx, commentID); err != nil {
				klog.Error("failed to delete comment: ", err)
			}
			continue
		}

		klog.Error("failed to unmarshal message: ", err)
	}
}

func CreateComment(ctx context.Context, comment *model.Comment) {
	value, err := json.Marshal(comment)
	if err != nil {
		klog.Error("failed to marshal message: ", err)
		return
	}
	commentMQInstance.Writer.WriteMessages(ctx, kafka.Message{
		Value: value,
	})
}

func DeleteComment(ctx context.Context, commentID int64) {
	value, err := json.Marshal(commentID)
	if err != nil {
		klog.Error("failed to marshal message: ", err)
		return
	}
	commentMQInstance.Writer.WriteMessages(ctx, kafka.Message{
		Value: value,
	})
}
