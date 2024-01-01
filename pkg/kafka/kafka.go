package kafka

import (
	"context"
	"douyin/config"
	"douyin/dal"
	"encoding/json"
	"log"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/segmentio/kafka-go"
)

const (
	topicDB = "DB"
	groupID = "backend"
)

type dbMessage struct {
	Data  map[string]string `json:"data"`
	Table string            `json:"table"`
	Type  string            `json:"type"`
}

// RemoveCache 删除redis缓存
func RemoveCache(ctx context.Context) {
	// 创建Reader
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  config.Conf.KafkaConfig.Brokers,
		Topic:    topicDB,
		GroupID:  groupID,
		MaxBytes: 10e6, // 10MB
	})

	// 接收消息
	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			klog.Error("failed to read message: ", err)
			break
		}
		msg := dbMessage{}
		if err := json.Unmarshal(m.Value, &msg); err != nil {
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
	if err := r.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}
}
