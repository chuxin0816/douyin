package kafka

import (
	"context"
	"encoding/json"

	"douyin/src/dal"

	"github.com/cloudwego/kitex/pkg/klog"
)

type syncMQ struct {
	*mq
}

var syncMQInstance *syncMQ

func initSyncMQ() {
	syncMQInstance = &syncMQ{
		mq: &mq{
			Topic:  topicCanal,
			Writer: NewWriter(topicCanal),
			Reader: NewReaderWithoutGroupID(topicCanal),
		},
	}
	go syncMQInstance.syncBloomFilter(context.Background())
}

func (mq *syncMQ) syncBloomFilter(ctx context.Context) {
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
		if msg.Type != "INSERT" {
			continue
		}
		if msg.Table == "user" {
			dal.AddToBloom(msg.Data[0]["id"])
			dal.AddToBloom(msg.Data[0]["name"])
		} else if msg.Table == "video" {
			dal.AddToBloom(msg.Data[0]["id"])
		}
	}

	// 程序退出前关闭Reader
	if err := mq.Reader.Close(); err != nil {
		klog.Fatal("failed to close reader:", err)
	}
}
