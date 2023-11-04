package middleware

import (
	"douyin/config"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	topic   = "douyin"
	groupID = "backend"
)

var (
	writer *kafka.Writer
	reader *kafka.Reader
)

func Init(conf *config.KafkaConfig) error {
	// 创建一个writer 向topic-A发送消息
	writer = &kafka.Writer{
		Addr:         kafka.TCP(conf.Brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{}, // 指定分区的balancer模式为最小字节分布
		RequiredAcks: kafka.RequireOne,    // ack模式
	}
	defer writer.Close()

	reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:        conf.Brokers,
		Topic:          topic,
		CommitInterval: 1 * time.Second,
		GroupID:        groupID,
	})
	defer reader.Close()
	return nil
}
