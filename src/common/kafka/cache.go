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
			Topic:  topicCanal,
			Writer: NewWriter(topicCanal),
			Reader: NewReader(topicCanal),
		},
	}
	go cacheMQInstance.removeCache(context.Background())
}

// removeCache 删除redis缓存
func (mq *cacheMQ) removeCache(ctx context.Context) {
	// 接收消息
	for {
		ctx, span := otel.Tracer("kafka").Start(ctx, "removeCache")

		m, err := mq.Reader.ReadMessage(ctx)
		if err != nil {
			klog.Error("failed to read message: ", err)
			break
		}
		msg := &dbMessage{}
		if err := json.Unmarshal(m.Value, msg); err != nil {
			klog.Error("failed to unmarshal message: ", err)
			continue
		}
		switch msg.Table {
		case "user":
			if msg.Type == "UPDATE" || msg.Type == "DELETE" {
				keyUserInfo := dal.GetRedisKey(dal.KeyUserInfoPF, msg.Data[0]["id"])
				dal.RDB.Del(ctx, keyUserInfo)
			}
		case "video":
			if msg.Type == "INSERT" {
				keyUserWorkCnt := dal.GetRedisKey(dal.KeyUserWorkCountPF, msg.Data[0]["author_id"])
				dal.RDB.Del(ctx, keyUserWorkCnt)
			} else if msg.Type == "UPDATE" {
				keyVideoInfo := dal.GetRedisKey(dal.KeyVideoInfoPF, msg.Data[0]["id"])
				dal.RDB.Del(ctx, keyVideoInfo)
			} else if msg.Type == "DELETE" {
				keyUserWorkCnt := dal.GetRedisKey(dal.KeyUserWorkCountPF, msg.Data[0]["author_id"])
				dal.RDB.Del(ctx, keyUserWorkCnt)
				keyVideoInfo := dal.GetRedisKey(dal.KeyVideoInfoPF, msg.Data[0]["id"])
				dal.RDB.Del(ctx, keyVideoInfo)
			}
		case "comment":
			if msg.Type == "INSERT" || msg.Type == "DELETE" {
				keyVideoCommentCnt := dal.GetRedisKey(dal.KeyVideoCommentCountPF, msg.Data[0]["video_id"])
				dal.RDB.Del(ctx, keyVideoCommentCnt)
			}
		}

		span.End()
	}

	// 程序退出前关闭Reader
	if err := mq.Reader.Close(); err != nil {
		klog.Fatal("failed to close reader:", err)
	}
}
