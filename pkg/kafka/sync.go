package kafka

import (
	"context"
	"douyin/dal"
	"douyin/dal/model"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/cloudwego/kitex/pkg/klog"
)

const (
	aggregateInterval = time.Second * 10
)

func syncRedisToMySQL() {
	ticker := time.NewTicker(aggregateInterval)
	defer ticker.Stop()
	for {
		<-ticker.C
		go syncUser()
		go syncVideo()
	}
}

func syncUser() {
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
		pipe.Get(context.Background(), dal.GetRedisKey(dal.KeyUserTotalFavoritedPF+userIDStr))
		pipe.Get(context.Background(), dal.GetRedisKey(dal.KeyUserFavoriteCountPF+userIDStr))
		pipe.Get(context.Background(), dal.GetRedisKey(dal.KeyUserFollowCountPF+userIDStr))
		pipe.Get(context.Background(), dal.GetRedisKey(dal.KeyUserFollowerCountPF+userIDStr))
		pipe.Get(context.Background(), dal.GetRedisKey(dal.KeyUserWorkCountPF+userIDStr))

		cmds, err := pipe.Exec(context.Background())
		if err != nil && err != redis.Nil {
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
		if err := UpdateUser(mUser); err != nil {
			klog.Error("同步redis用户缓存到mysql失败,err: ", err)
			continue
		}
	}
}
func syncVideo() {
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
		pipe.Get(context.Background(), dal.GetRedisKey(dal.KeyVideoFavoriteCountPF+videoIDStr))
		pipe.Get(context.Background(), dal.GetRedisKey(dal.KeyVideoCommentCountPF+videoIDStr))

		cmds, err := pipe.Exec(context.Background())
		if err != nil {
			if err == redis.Nil {
				klog.Warnf("redis中不存在视频ID为%d的缓存", videoID)
				continue
			}
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
		if err := UpdateVideo(mVideo); err != nil {
			klog.Errorf("同步redis视频缓存到mysql失败,err: ", err)
			continue
		}
	}
}
