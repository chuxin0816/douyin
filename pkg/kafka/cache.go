package kafka

import (
	"context"
	"douyin/dal"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/kitex/pkg/klog"
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
	go cacheMQInstance.removeCache(context.Background())
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
			fmt.Println("err:", err)
			klog.Error("failed to unmarshal message: ", err)
			continue
		}
		if msg.Type != "DELETE" {
			continue
		}
		switch msg.Table {
		case "favorite":
			if err := dal.RemoveFavoriteCache(ctx, msg.Data[0]["user_id"], msg.Data[0]["video_id"]); err != nil {
				klog.Error("failed to remove favorite cache:", err)
			}
		case "relation":
			if err := dal.RemoveRelationCache(ctx, msg.Data[0]["follower_id"], msg.Data[0]["user_id"]); err != nil {
				klog.Error("failed to remove relation cache:", err)
			}
		}
	}

	// 程序退出前关闭Reader
	if err := mq.Reader.Close(); err != nil {
		klog.Fatal("failed to close reader:", err)
	}
}
