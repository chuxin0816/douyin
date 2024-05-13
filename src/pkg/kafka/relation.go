package kafka

import (
	"context"
	"encoding/json"

	"douyin/src/dal"
	"douyin/src/dal/model"
	"douyin/src/pkg/tracing"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/codes"
)

type relationMQ struct {
	*mq
}

var relationMQInstance *relationMQ

func initRelationMQ() {
	relationMQInstance = &relationMQ{
		&mq{
			Topic:  topicRelation,
			Writer: NewWriter(topicRelation),
			Reader: NewReader(topicRelation),
		},
	}

	go relationMQInstance.consumeRelation(context.Background())
}

func (mq *relationMQ) consumeRelation(ctx context.Context) {
	// 接收消息
	for {
		m, err := mq.Reader.ReadMessage(ctx)
		if err != nil {
			klog.Error("failed to read message: ", err)
			break
		}

		relation := &model.Relation{}
		if err := json.Unmarshal(m.Value, relation); err != nil {
			klog.Error("failed to unmarshal message: ", err)
			continue
		}

		// 检查记录是否存在
		exist, err := dal.CheckRelationExist(ctx, relation.FollowerID, relation.AuthorID)
		if err != nil {
			klog.Error("检查记录是否存在失败, err: ", err)
			continue
		}

		if exist {
			// 存在则取关
			if err := dal.UnFollow(ctx, relation.FollowerID, relation.AuthorID); err != nil {
				klog.Error("删除记录失败, err: ", err)
				continue
			}
		} else {
			// 不存在则关注
			if err := dal.Follow(ctx, relation.FollowerID, relation.AuthorID); err != nil {
				klog.Error("添加记录失败, err: ", err)
				continue
			}
		}
	}

	// 程序退出前关闭Reader
	if err := mq.Reader.Close(); err != nil {
		klog.Fatal("failed to close reader:", err)
	}
}

func Relation(ctx context.Context, relation *model.Relation) error {
	ctx, span := tracing.Tracer.Start(ctx, "kafka.Relation")
	defer span.End()

	data, err := json.Marshal(relation)
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