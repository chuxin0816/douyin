package kafka

import (
	"context"
	"encoding/json"

	"douyin/src/dal"

	"github.com/cloudwego/kitex/pkg/klog"
	"go.opentelemetry.io/otel"
)

type cacheMQ struct {
	*mq
}

var cacheMQInstance *cacheMQ

func initCacheMQ() {
	cacheMQInstance = &cacheMQ{
		mq: &mq{
			Topic:  topicDebezium,
			Writer: NewWriter(topicDebezium),
			Reader: NewReader(topicDebezium),
		},
	}
	go cacheMQInstance.removeCache(context.Background())
}

// removeCache 删除redis缓存
func (mq *cacheMQ) removeCache(ctx context.Context) {
	// 接收消息
	for {
		ctx, span := otel.Tracer("kafka").Start(ctx, "removeCache")

		m, err := mq.Reader.FetchMessage(ctx)
		if err != nil {
			klog.Error("failed to fetch message: ", err)
			span.End()
			break
		}
		msg := &dbMessage{}
		if err := json.Unmarshal(m.Value, msg); err != nil {
			klog.Error("failed to unmarshal message: ", err)
			span.End()
			continue
		}

		pipe := dal.RDB.Pipeline()
		switch msg.Table {
		case "user":
			if msg.Type == "UPDATE" || msg.Type == "DELETE" {
				keyUserInfo := dal.GetRedisKey(dal.KeyUserInfoPF, msg.Data[0]["id"])
				pipe.Del(ctx, keyUserInfo)
			}
		case "video":
			if msg.Type == "INSERT" {
				keyUserWorkCnt := dal.GetRedisKey(dal.KeyUserWorkCountPF, msg.Data[0]["author_id"])
				dal.IncrByScript.Run(ctx, pipe, []string{keyUserWorkCnt}, 1)
			} else if msg.Type == "UPDATE" {
				keyVideoInfo := dal.GetRedisKey(dal.KeyVideoInfoPF, msg.Data[0]["id"])
				pipe.Del(ctx, keyVideoInfo)
			} else if msg.Type == "DELETE" {
				keyUserWorkCnt := dal.GetRedisKey(dal.KeyUserWorkCountPF, msg.Data[0]["author_id"])
				dal.IncrByScript.Run(ctx, pipe, []string{keyUserWorkCnt}, -1)
				keyVideoInfo := dal.GetRedisKey(dal.KeyVideoInfoPF, msg.Data[0]["id"])
				pipe.Del(ctx, keyVideoInfo)
			}
		case "comment":
			if msg.Type == "INSERT" {
				keyVideoCommentCnt := dal.GetRedisKey(dal.KeyVideoCommentCountPF, msg.Data[0]["video_id"])
				dal.IncrByScript.Run(ctx, pipe, []string{keyVideoCommentCnt}, 1)
			} else if msg.Type == "DELETE" {
				keyVideoCommentCnt := dal.GetRedisKey(dal.KeyVideoCommentCountPF, msg.Data[0]["video_id"])
				dal.IncrByScript.Run(ctx, pipe, []string{keyVideoCommentCnt}, -1)
			}
		}

		if _, err := pipe.Exec(ctx); err != nil {
			klog.Error("failed to update or delete cache: ", err)
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
