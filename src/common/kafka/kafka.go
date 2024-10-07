package kafka

import (
	"time"

	"douyin/src/config"

	"github.com/segmentio/kafka-go"
)

const (
	topicDebezium  = "debezium"
	topicComment   = "comment"
	topicFavorite  = "favorite"
	topicRelation  = "relation"
	topicMessage   = "message"
	groupID        = "backend"
	commitInterval = 1 * time.Second
)

type mq struct {
	Topic  string
	Writer *kafka.Writer
	Reader *kafka.Reader
}

type dbMessage struct {
	Data  []map[string]string `json:"data"`
	Table string              `json:"table"`
	Type  string              `json:"type"`
}

func Init() {
	initCacheMQ()
	initCommentMQ()
	initFavoriteMQ()
	initMessageMQ()
	initRelationMQ()
}

func NewWriter(topic string) *kafka.Writer {
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers:      config.Conf.KafkaConfig.Brokers,
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: int(kafka.RequireOne),
		MaxAttempts:  3,
	})
}

func NewReader(topic string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:        config.Conf.KafkaConfig.Brokers,
		Topic:          topic,
		GroupID:        groupID,
		MaxBytes:       10e6,           // 10MB
		CommitInterval: commitInterval, // 每秒刷新一次提交给 Kafka
	})
}
