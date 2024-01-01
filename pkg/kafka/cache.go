package kafka

import (
	"context"
	"douyin/dal"
	"encoding/json"

	"github.com/u2takey/go-utils/klog"
)

type cacheMQ struct {
	*mq
}

var cacheMQInstance *cacheMQ

func initCacheMQ() {
	cacheMQInstance = &cacheMQ{
		mq: &mq{
			Topic:  topicCache,
			Writer: NewWriter(topicCache),
			Reader: NewReader(topicCache),
		},
	}

	ctx := context.Background()
	go cacheMQInstance.removeCache(ctx)
}

// removeCache 删除redis缓存
func (mq *cacheMQ) removeCache(ctx context.Context) {
	// 接收消息
	for {
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
		case "favorites":
			dal.RemoveFavoriteCache(ctx, msg.Data["user_id"], msg.Data["video_id"])
		case "relations":
			dal.RemoveRelationCache(ctx, msg.Data["follower_id"], msg.Data["user_id"])
		}
	}

	// 程序退出前关闭Reader
	if err := mq.Reader.Close(); err != nil {
		klog.Fatal("failed to close reader:", err)
	}
}
