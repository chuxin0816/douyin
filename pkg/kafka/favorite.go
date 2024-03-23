package kafka

import (
	"context"
	"douyin/dal"
	"douyin/dal/model"
	"douyin/pkg/tracing"
	"encoding/json"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/codes"
)

type favoriteMQ struct {
	*mq
}

var favoriteMQInstance *favoriteMQ

func initFavoriteMQ() {
	favoriteMQInstance = &favoriteMQ{
		&mq{
			Topic:  topicFavorite,
			Writer: NewWriter(topicFavorite),
			Reader: NewReader(topicFavorite),
		},
	}

	go favoriteMQInstance.consumeFavorite(context.Background())
}

func (mq *favoriteMQ) consumeFavorite(ctx context.Context) {
	ctx, span := tracing.Tracer.Start(ctx, "kafka.consumeFavorite")
	defer span.End()

	// 接收消息
	for {
		m, err := mq.Reader.ReadMessage(ctx)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to read message")
			klog.Error("failed to read message: ", err)
			break
		}

		favorite := &model.Favorite{}
		if err := json.Unmarshal(m.Value, favorite); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to unmarshal message")
			klog.Error("failed to unmarshal message: ", err)
			continue
		}

		// 检查记录是否存在
		exist, err := dal.CheckFavoriteExist(ctx, favorite.UserID, favorite.VideoID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to check record exist")
			klog.Error("检查记录是否存在失败, err: ", err)
			continue
		}

		if exist {
			// 存在则删除
			if err := dal.DeleteFavorite(ctx, favorite.UserID, favorite.VideoID); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "failed to delete record")
				klog.Error("删除记录失败, err: ", err)
				continue
			}
		} else {
			// 不存在则添加
			if err := dal.CreateFavorite(ctx, favorite.UserID, favorite.VideoID); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "failed to add record")
				klog.Error("添加记录失败, err: ", err)
				continue
			}
		}
	}
	// 程序退出前关闭Reader
	if err := mq.Reader.Close(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to close reader")
		klog.Fatal("failed to close reader:", err)
	}
}

func Favorite(favorite *model.Favorite) error {
	_, span := tracing.Tracer.Start(context.Background(), "kafka.Favorite")
	defer span.End()

	data, err := json.Marshal(favorite)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal message")
		klog.Error("failed to marshal message: ", err)
		return err
	}

	return favoriteMQInstance.Writer.WriteMessages(context.Background(), kafka.Message{
		Value: data,
	})
}
