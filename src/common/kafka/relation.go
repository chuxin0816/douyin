package kafka

import (
	"context"
	"strconv"

	"douyin/src/dal"
	"douyin/src/dal/model"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/segmentio/kafka-go"
	"github.com/vmihailenco/msgpack/v5"
	"go.opentelemetry.io/otel"
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
		ctx, span := otel.Tracer("kafka").Start(ctx, "consumeRelation")

		m, err := mq.Reader.FetchMessage(ctx)
		if err != nil {
			klog.Error("failed to fetch message: ", err)
			span.End()
			break
		}

		relation := &model.Relation{}
		if err := msgpack.Unmarshal(m.Value, relation); err != nil {
			klog.Error("failed to unmarshal message: ", err)
			span.End()
			continue
		}

		pipe := dal.RDB.Pipeline()

		if relation.ID == 1 {
			// 关注
			if err := dal.Follow(ctx, relation.FollowerID, relation.AuthorID); err != nil {
				klog.Error("添加记录失败, err: ", err)
				span.End()
				continue
			}
			dal.IncrByScript.Run(ctx, pipe, []string{dal.GetRedisKey(dal.KeyUserFollowCountPF, strconv.FormatInt(relation.FollowerID, 10))}, 1)
			dal.IncrByScript.Run(ctx, pipe, []string{dal.GetRedisKey(dal.KeyUserFollowerCountPF, strconv.FormatInt(relation.FollowerID, 10))}, 1)
			pipe.SAdd(ctx, dal.GetRedisKey(dal.KeyUserFollowPF, strconv.FormatInt(relation.FollowerID, 10)), relation.AuthorID)
			pipe.SAdd(ctx, dal.GetRedisKey(dal.KeyUserFollowerPF, strconv.FormatInt(relation.AuthorID, 10)), relation.FollowerID)
			if exist, err := dal.CheckRelationExist(ctx, relation.AuthorID, relation.FollowerID); err != nil {
				klog.Error("查询关注记录失败, err: ", err)
			} else if exist {
				pipe.SAdd(ctx, dal.GetRedisKey(dal.KeyUserFriendPF, strconv.FormatInt(relation.FollowerID, 10)), relation.AuthorID)
				pipe.SAdd(ctx, dal.GetRedisKey(dal.KeyUserFriendPF, strconv.FormatInt(relation.AuthorID, 10)), relation.FollowerID)
			}
		} else {
			// 取关
			if err := dal.UnFollow(ctx, relation.FollowerID, relation.AuthorID); err != nil {
				klog.Error("删除记录失败, err: ", err)
				span.End()
				continue
			}
			dal.IncrByScript.Run(ctx, pipe, []string{dal.GetRedisKey(dal.KeyUserFollowCountPF, strconv.FormatInt(relation.FollowerID, 10))}, -1)
			dal.IncrByScript.Run(ctx, pipe, []string{dal.GetRedisKey(dal.KeyUserFollowerCountPF, strconv.FormatInt(relation.FollowerID, 10))}, -1)
			pipe.SRem(ctx, dal.GetRedisKey(dal.KeyUserFollowPF, strconv.FormatInt(relation.FollowerID, 10)), relation.AuthorID)
			pipe.SRem(ctx, dal.GetRedisKey(dal.KeyUserFollowerPF, strconv.FormatInt(relation.AuthorID, 10)), relation.FollowerID)
			if exist, err := dal.CheckRelationExist(ctx, relation.AuthorID, relation.FollowerID); err != nil {
				klog.Error("查询关注记录失败, err: ", err)
			} else if exist {
				pipe.SRem(ctx, dal.GetRedisKey(dal.KeyUserFriendPF, strconv.FormatInt(relation.FollowerID, 10)), relation.AuthorID)
				pipe.SRem(ctx, dal.GetRedisKey(dal.KeyUserFriendPF, strconv.FormatInt(relation.AuthorID, 10)), relation.FollowerID)
			}
		}

		// 更新缓存
		if _, err = pipe.Exec(ctx); err != nil {
			klog.Error("删除缓存失败, err: ", err)
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

func Relation(ctx context.Context, relation *model.Relation) error {
	ctx, span := otel.Tracer("kafka").Start(ctx, "kafka.Relation")
	defer span.End()

	data, err := msgpack.Marshal(relation)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal message")
		klog.Error("failed to marshal message: ", err)
		return err
	}

	return relationMQInstance.Writer.WriteMessages(ctx, kafka.Message{
		Value: data,
	})
}
