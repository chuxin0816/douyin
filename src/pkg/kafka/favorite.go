package kafka

import (
	"context"
	"sync"
	"time"

	"douyin/src/dal"
	"douyin/src/dal/model"
	"douyin/src/pkg/snowflake"
	"douyin/src/pkg/tracing"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/segmentio/kafka-go"
	"github.com/vmihailenco/msgpack/v5"
	"go.opentelemetry.io/otel/codes"
)

type favoriteMQ struct {
	*mq
}

const syncInterval = time.Second * 5

var (
	favoriteMQInstance *favoriteMQ
	FavoriteMap        = make(map[int64]map[int64]int64)
	Mu                 sync.Mutex
)

func initFavoriteMQ() {
	favoriteMQInstance = &favoriteMQ{
		&mq{
			Topic:  topicFavorite,
			Writer: NewWriter(topicFavorite),
			Reader: NewReader(topicFavorite),
		},
	}

	go favoriteMQInstance.consumeFavorite(context.Background())
	go syncFavoriteToDB()
}

func (mq *favoriteMQ) consumeFavorite(ctx context.Context) {
	// 接收消息
	for {
		m, err := mq.Reader.ReadMessage(ctx)
		if err != nil {
			klog.Error("failed to read message: ", err)
			break
		}

		favorites := make([]*model.Favorite, 0)
		if err := msgpack.Unmarshal(m.Value, &favorites); err != nil {
			klog.Error("failed to unmarshal message: ", err)
			continue
		}

		creates := make([]*model.Favorite, 0, len(favorites))
		deletes := make([]int64, 0, len(favorites))

		for _, favorite := range favorites {
			if favorite.ID == 1 {
				favorite.ID = snowflake.GenerateID()
				creates = append(creates, favorite)
			} else {
				deletes = append(deletes, favorite.UserID)
			}
		}

		if err := dal.BatchCreateFavorite(ctx, creates); err != nil {
			klog.Error("failed to create favorite: ", err)
			continue
		}
		if err := dal.BatchDeleteFavorite(ctx, deletes); err != nil {
			klog.Error("failed to delete favorite: ", err)
			continue
		}
	}

	// 程序退出前关闭Reader
	if err := mq.Reader.Close(); err != nil {
		klog.Fatal("failed to close reader:", err)
	}
}

func BatchFavorite(ctx context.Context, favorites []*model.Favorite) error {
	ctx, span := tracing.Tracer.Start(ctx, "kafka.Favorite")
	defer span.End()

	data, err := msgpack.Marshal(favorites)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal message")
		klog.Error("failed to marshal message: ", err)
		return err
	}

	return favoriteMQInstance.Writer.WriteMessages(ctx, kafka.Message{
		Value: data,
	})
}

// syncFavoriteToDB 同步点赞数据到数据库
func syncFavoriteToDB() {
	ticker := time.NewTicker(syncInterval)
	for range ticker.C {
		Mu.Lock()
		backupFavoriteMap := FavoriteMap
		FavoriteMap = make(map[int64]map[int64]int64)
		Mu.Unlock()

		favorites := make([]*model.Favorite, 0, len(backupFavoriteMap))
		for userID, videoMap := range backupFavoriteMap {
			for videoID, actionType := range videoMap {
				favorites = append(favorites, &model.Favorite{
					ID:      actionType,
					UserID:  userID,
					VideoID: videoID,
				})
			}
		}

		if err := BatchFavorite(context.Background(), favorites); err != nil {
			klog.Error("通过kafka更新favorite表失败, err: ", err)
			continue
		}

	}
}
