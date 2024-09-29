package kafka

import (
	"context"
	"strconv"
	"strings"
	"time"

	"douyin/src/common/snowflake"
	"douyin/src/common/utils"
	"douyin/src/dal"
	"douyin/src/dal/model"

	cmap "github.com/chuxin0816/concurrent-map"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/segmentio/kafka-go"
	"github.com/vmihailenco/msgpack/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type favoriteMQ struct {
	*mq
}

type FavoriteEvent struct {
	UserID     int64
	VideoID    int64
	ActionType int64
}

// 使用 FNV-1a 生成唯一的较短字符串
func (f FavoriteEvent) String() string {
	// 拼接三个 int64 字段并写入哈希
	var builder strings.Builder
	builder.WriteString(strconv.FormatInt(f.UserID, 10))
	builder.WriteString("_")
	builder.WriteString(strconv.FormatInt(f.VideoID, 10))
	builder.WriteString("_")
	builder.WriteString(strconv.FormatInt(f.ActionType, 10))

	return strconv.FormatUint(uint64(utils.Fnv32a(builder.String())), 10)
}

const syncInterval = time.Second * 10

var (
	favoriteMQInstance *favoriteMQ
	FavoriteMap        *cmap.ConcurrentMap[FavoriteEvent]
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
	go syncFavoriteToDB(context.Background())

	FavoriteMap = cmap.New(cmap.WithShardCount[FavoriteEvent](128))
}

func (mq *favoriteMQ) consumeFavorite(ctx context.Context) {
	// 接收消息
	for {
		ctx, span := otel.Tracer("kafka").Start(ctx, "consumeFavorite")

		m, err := mq.Reader.FetchMessage(ctx)
		if err != nil {
			klog.Error("failed to fetch message: ", err)
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

		if err := mq.Reader.CommitMessages(ctx, m); err != nil {
			klog.Error("failed to commit message: ", err)
		}

		span.End()
	}

	// 程序退出前关闭Reader
	if err := mq.Reader.Close(); err != nil {
		klog.Fatal("failed to close reader:", err)
	}
}

func BatchFavorite(ctx context.Context, favorites []*model.Favorite) error {
	ctx, span := otel.Tracer("kafka").Start(ctx, "kafka.Favorite")
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
func syncFavoriteToDB(ctx context.Context) {
	ticker := time.NewTicker(syncInterval)
	for range ticker.C {
		_, span := otel.Tracer("kafka").Start(ctx, "syncFavoriteToDB")

		events := FavoriteMap.PopAll()
		favorites := make([]*model.Favorite, 0, len(events))
		for event := range events {
			favorites = append(favorites, &model.Favorite{
				ID:      event.Val.ActionType,
				UserID:  event.Val.UserID,
				VideoID: event.Val.VideoID,
			})
		}

		if err := BatchFavorite(context.Background(), favorites); err != nil {
			klog.Error("通过kafka更新favorite表失败, err: ", err)
			continue
		}

		span.End()
	}
}
