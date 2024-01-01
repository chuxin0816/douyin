package dal

import (
	"context"
	"douyin/dal/model"
	"strconv"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

func FavoriteAction(userID int64, videoID int64, actionType int64) error {
	// 先判断布隆过滤器中是否存在
	if !bloomFilter.Test([]byte(strconv.FormatInt(videoID, 10))) {
		klog.Error("视频不存在, videoID: ", videoID)
		return ErrVideoNotExist
	}

	// 查看是否已经点赞
	key := GetRedisKey(KeyUserFavoritePF + strconv.FormatInt(userID, 10))
	// 使用singleflight避免缓存击穿和减少缓存压力
	_, err, _ := g.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(delayTime)
			g.Forget(key)
		}()
		exist := RDB.SIsMember(context.Background(), key, videoID).Val()
		if exist {
			if actionType == 1 {
				return nil, ErrAlreadyFavorite
			}
			return nil, nil
		}
		// 缓存未命中，查询mysql中是否有记录
		var id int64
		if err := qFavorite.WithContext(context.Background()).Where(qFavorite.UserID.Eq(userID), qFavorite.VideoID.Eq(videoID)).Select(qFavorite.ID).Scan(&id); err != nil {
			klog.Error("查询mysql中是否有记录失败, err: ", err)
			return nil, err
		}
		// mysql中有记录
		if id != 0 && actionType == 1 {
			// 写入redis缓存
			go func() {
				if err := RDB.SAdd(context.Background(), key, videoID).Err(); err != nil {
					klog.Error("写入redis缓存失败, err: ", err)
					return
				}
				if err := RDB.Expire(context.Background(), key, expireTime+getRandomTime()).Err(); err != nil {
					klog.Error("设置redis缓存过期时间失败, err: ", err)
					return
				}
			}()
			return nil, ErrAlreadyFavorite
		}
		// mysql中没有记录
		if id == 0 && actionType == -1 {
			return nil, ErrNotFavorite
		}
		return nil, nil
	})
	if err != nil {
		return err
	}

	// 先查询作者的ID
	var authorID int64
	if err = qVideo.WithContext(context.Background()).Where(qVideo.ID.Eq(videoID)).Select(qVideo.AuthorID).Scan(&authorID); err != nil {
		klog.Error("查询作者的ID失败, err: ", err)
		return err
	}

	// 保存用户点赞视频的记录, 采用延迟双删策略
	// 删除redis缓存
	if err := RDB.SRem(context.Background(), key, videoID).Err(); err != nil {
		klog.Error("删除redis缓存失败, err: ", err)
	}

	// 更新favorite表
	if actionType == 1 {
		if err := qFavorite.WithContext(context.Background()).Create(&model.Favorite{UserID: userID, VideoID: videoID}); err != nil {
			klog.Error("更新favorite表失败, err: ", err)
			return err
		}
	} else {
		if _, err := qFavorite.WithContext(context.Background()).Where(qFavorite.UserID.Eq(userID), qFavorite.VideoID.Eq(videoID)).Delete(); err != nil {
			klog.Error("更新favorite表失败, err: ", err)
			return err
		}
	}
	// 延迟后删除redis缓存, 由kafka任务处理

	// 更新video的favorite_count字段
	if err := RDB.IncrBy(context.Background(), GetRedisKey(KeyVideoFavoriteCountPF+strconv.FormatInt(videoID, 10)), actionType).Err(); err != nil {
		klog.Error("更新video的favorite_count字段失败, err: ", err)
		return err
	}

	// 更新user当前用户的favorite_count字段
	if err := RDB.IncrBy(context.Background(), GetRedisKey(KeyUserFavoriteCountPF+strconv.FormatInt(userID, 10)), actionType).Err(); err != nil {
		klog.Error("更新user当前用户的favorite_count字段失败, err: ", err)
		return err
	}

	// 更新user作者的total_favorited字段
	if err := RDB.IncrBy(context.Background(), GetRedisKey(KeyUserTotalFavoritedPF+strconv.FormatInt(authorID, 10)), actionType).Err(); err != nil {
		klog.Error("更新user作者的total_favorited字段失败, err: ", err)
		return err
	}

	// 写入待同步切片
	CacheUserID.Store(userID, struct{}{})
	CacheUserID.Store(authorID, struct{}{})
	CacheVideoID.Store(videoID, struct{}{})

	return nil
}

func GetFavoriteList(userID int64) ([]int64, error) {
	// 先查询redis缓存
	key := GetRedisKey(KeyUserFavoritePF + strconv.FormatInt(userID, 10))
	videoIDStrs := RDB.SMembers(context.Background(), key).Val()
	if len(videoIDStrs) != 0 {
		videoIDs := make([]int64, 0, len(videoIDStrs))
		for _, videoIDStr := range videoIDStrs {
			videoID, _ := strconv.ParseInt(videoIDStr, 10, 64)
			videoIDs = append(videoIDs, videoID)
		}
		return videoIDs, nil
	}

	// 缓存未命中 ,查询mysql, 使用singleflight防止缓存击穿
	var videoIDs []int64
	_, err, _ := g.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(delayTime)
			g.Forget(key)
		}()
		if err := qFavorite.WithContext(context.Background()).Where(qFavorite.UserID.Eq(userID)).Select(qFavorite.VideoID).Scan(&videoIDs); err != nil {
			klog.Error("查询favorite表失败, err: ", err)
			return nil, err
		}

		// 写入redis缓存
		if len(videoIDs) > 0 {
			go func() {
				pipeline := RDB.Pipeline()
				for _, videoID := range videoIDs {
					pipeline.SAdd(context.Background(), key, videoID)
				}
				pipeline.Expire(context.Background(), key, expireTime+getRandomTime())
				if _, err := pipeline.Exec(context.Background()); err != nil {
					klog.Error("写入redis缓存失败, err: ", err)
				}
			}()
		}
		return videoIDs, nil
	})
	if err != nil {
		return nil, err
	}
	return videoIDs, nil
}
