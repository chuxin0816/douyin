package kafka

import (
	"context"
	"douyin/dal"
	"douyin/dal/model"
	"douyin/pkg/tracing"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/codes"

	"github.com/cloudwego/kitex/pkg/klog"
)

const (
	aggregateInterval = time.Second * 10
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
			fmt.Println("err:", err)
			klog.Error("failed to unmarshal message: ", err)
			continue
		}
		if msg.Type != "INSERT" {
			continue
		}
		if msg.Table == "user" || msg.Table == "video" {
			dal.AddToBloom(msg.Data[0]["id"])
		}
	}

	// 程序退出前关闭Reader
	if err := mq.Reader.Close(); err != nil {
		klog.Fatal("failed to close reader:", err)
	}
}

func syncRedisToMySQL(ctx context.Context) {
	ticker := time.NewTicker(aggregateInterval)
	defer ticker.Stop()
	for {
		<-ticker.C
		go syncUser(ctx)
		go syncVideo(ctx)
	}
}

func syncUser(ctx context.Context) {
	ctx, span := tracing.Tracer.Start(ctx, "kafka.syncUser")
	defer span.End()

	// 备份缓存中的用户ID并清空
	backupUserID := make([]int64, 0, 100000)

	dal.CacheUserID.Range(func(key, value any) bool {
		backupUserID = append(backupUserID, key.(int64))
		dal.CacheUserID.Delete(key)
		return true
	})

	// 同步redis的用户缓存到Mysql
	pipe := dal.RDB.Pipeline()

	for _, userID := range backupUserID {
		userIDStr := strconv.FormatInt(userID, 10)
		pipe.Get(ctx, dal.GetRedisKey(dal.KeyUserTotalFavoritedPF+userIDStr))
		pipe.Get(ctx, dal.GetRedisKey(dal.KeyUserFavoriteCountPF+userIDStr))
		pipe.Get(ctx, dal.GetRedisKey(dal.KeyUserFollowCountPF+userIDStr))
		pipe.Get(ctx, dal.GetRedisKey(dal.KeyUserFollowerCountPF+userIDStr))
		pipe.Get(ctx, dal.GetRedisKey(dal.KeyUserWorkCountPF+userIDStr))

		cmds, err := pipe.Exec(ctx)
		if err != nil && err != redis.Nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to exec pipeline")
			klog.Error("同步redis用户缓存到mysql失败,err: ", err)
			return
		}

		totalFavorited, _ := strconv.ParseInt(cmds[0].(*redis.StringCmd).Val(), 10, 64)
		favoriteCount, _ := strconv.ParseInt(cmds[1].(*redis.StringCmd).Val(), 10, 64)
		followCount, _ := strconv.ParseInt(cmds[2].(*redis.StringCmd).Val(), 10, 64)
		followerCount, _ := strconv.ParseInt(cmds[3].(*redis.StringCmd).Val(), 10, 64)
		workCount, _ := strconv.ParseInt(cmds[4].(*redis.StringCmd).Val(), 10, 64)
		mUser := &model.User{
			ID:             userID,
			TotalFavorited: totalFavorited,
			FavoriteCount:  favoriteCount,
			FollowCount:    followCount,
			FollowerCount:  followerCount,
			WorkCount:      workCount,
		}
		if err := UpdateUser(ctx, mUser); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "同步redis用户缓存到mysql失败")
			klog.Error("同步redis用户缓存到mysql失败,err: ", err)
			continue
		}
	}
}
func syncVideo(ctx context.Context) {
	ctx, span := tracing.Tracer.Start(ctx, "kafka.syncVideo")
	defer span.End()

	// 备份缓存中的视频ID并清空
	backupVideoID := make([]int64, 0, 100000)

	dal.CacheVideoID.Range(func(key, value any) bool {
		backupVideoID = append(backupVideoID, key.(int64))
		dal.CacheVideoID.Delete(key)
		return true
	})

	// 同步redis中的视频缓存到Mysql
	pipe := dal.RDB.Pipeline()
	for _, videoID := range backupVideoID {
		videoIDStr := strconv.FormatInt(videoID, 10)
		pipe.Get(ctx, dal.GetRedisKey(dal.KeyVideoFavoriteCountPF+videoIDStr))
		pipe.Get(ctx, dal.GetRedisKey(dal.KeyVideoCommentCountPF+videoIDStr))

		cmds, err := pipe.Exec(ctx)
		if err != nil && err != redis.Nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "同步redis视频缓存到mysql失败")
			klog.Errorf("同步redis视频缓存到mysql失败,err: ", err)
			continue
		}

		videoFavoriteCount, _ := strconv.ParseInt(cmds[0].(*redis.StringCmd).Val(), 10, 64)
		videoCommentCount, _ := strconv.ParseInt(cmds[1].(*redis.StringCmd).Val(), 10, 64)
		mVideo := &model.Video{
			ID:            videoID,
			FavoriteCount: videoFavoriteCount,
			CommentCount:  videoCommentCount,
		}
		if err := UpdateVideo(ctx, mVideo); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "同步redis视频缓存到mysql失败")
			klog.Errorf("同步redis视频缓存到mysql失败,err: ", err)
			continue
		}
	}
}
