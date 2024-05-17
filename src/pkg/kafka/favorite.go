package kafka

import (
	"context"
	"strconv"
	"sync"
	"time"

	"douyin/src/dal"
	"douyin/src/dal/model"
	"douyin/src/pkg/tracing"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/segmentio/kafka-go"
	"github.com/vmihailenco/msgpack/v5"
	"go.opentelemetry.io/otel/codes"
)

type favoriteMQ struct {
	*mq
}

const syncInterval = time.Second * 2

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

		favorite := &model.Favorite{}
		if err := msgpack.Unmarshal(m.Value, favorite); err != nil {
			klog.Error("failed to unmarshal message: ", err)
			continue
		}

		if favorite.ID == 1 {
			// 点赞
			if err := dal.CreateFavorite(ctx, favorite.UserID, favorite.VideoID); err != nil {
				klog.Error("添加记录失败, err: ", err)
				continue
			}
		} else {
			// 取消点赞
			if err := dal.DeleteFavorite(ctx, favorite.UserID, favorite.VideoID); err != nil {
				klog.Error("删除记录失败, err: ", err)
				continue
			}
		}

		// 删除缓存
		pipe := dal.RDB.Pipeline()
		pipe.Del(ctx, dal.GetRedisKey(dal.KeyUserFollowPF, strconv.FormatInt(favorite.UserID, 10)))
		pipe.Del(ctx, dal.GetRedisKey(dal.KeyUserFollowerPF, strconv.FormatInt(favorite.UserID, 10)))
		pipe.Del(ctx, dal.GetRedisKey(dal.KeyUserFriendPF, strconv.FormatInt(favorite.UserID, 10)))
		pipe.Del(ctx, dal.GetRedisKey(dal.KeyUserFollowCountPF, strconv.FormatInt(favorite.UserID, 10)))
		pipe.Del(ctx, dal.GetRedisKey(dal.KeyUserFollowerCountPF, strconv.FormatInt(favorite.UserID, 10)))
		_, err = pipe.Exec(ctx)
		if err != nil {
			klog.Error("删除缓存失败, err: ", err)
		}
	}

	// 程序退出前关闭Reader
	if err := mq.Reader.Close(); err != nil {
		klog.Fatal("failed to close reader:", err)
	}
}

func Favorite(ctx context.Context, favorite *model.Favorite) error {
	ctx, span := tracing.Tracer.Start(ctx, "kafka.Favorite")
	defer span.End()

	data, err := msgpack.Marshal(favorite)
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

		for userID, videoMap := range backupFavoriteMap {
			for videoID, actionType := range videoMap {
				// 通过kafka更新favorite表
				err := Favorite(context.Background(), &model.Favorite{
					ID:      actionType,
					UserID:  userID,
					VideoID: videoID,
				})
				if err != nil {
					klog.Error("通过kafka更新favorite表失败, err: ", err)
					continue
				}
			}
		}
	}
}
