package kafka

import (
	"douyin/src/config"

	"github.com/segmentio/kafka-go"
)

const (
	topicCanal    = "canal"
	topicComment  = "comment"
	topicFavorite = "favorite"
	topicRelation = "relation"
	topicMessage  = "message"
	groupID       = "backend"
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
	})
}

func NewReader(topic string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  config.Conf.KafkaConfig.Brokers,
		Topic:    topic,
		GroupID:  groupID,
		MaxBytes: 10e6, // 10MB
	})
}