package dao

import (
	"context"
	"douyin/models"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func FavoriteAction(userID int64, videoID int64, actionType int64) error {
	// 先判断布隆过滤器中是否存在
	if !bloomFilter.Test([]byte(strconv.FormatInt(videoID, 10))) {
		hlog.Error("mysql.FavoriteAction: 视频不存在, videoID: ", videoID)
		return ErrVideoNotExist
	}

	// 查看是否已经点赞
	key := getRedisKey(KeyVideoFavoritePF + strconv.FormatInt(videoID, 10))
	exist := rdb.SIsMember(context.Background(), key, userID).Val()
	if exist && actionType == 1 {
		return ErrAlreadyFavorite
	}
	// 缓存未命中，查询mysql中是否有记录
	if !exist {
		// 使用singleflight避免缓存击穿
		_, err, _ := g.Do(key, func() (interface{}, error) {
			go func() {
				time.Sleep(delayTime)
				g.Forget(key)
			}()
			var id int64
			if err := db.Model(&models.Favorite{}).Where("user_id = ? AND video_id = ?", userID, videoID).Select("id").Scan(&id).Error; err != nil {
				hlog.Error("mysql.FavoriteAction: 查询mysql中是否有记录失败, err: ", err)
				return nil, err
			}
			// mysql中有记录
			if id != 0 && actionType == 1 {
				// 写入redis缓存
				if err := rdb.SAdd(context.Background(), key, userID).Err(); err != nil {
					hlog.Error("redis.FavoriteAction: 写入redis缓存失败, err: ", err)
					return nil, err
				}
				if err := rdb.Expire(context.Background(), key, expireTime+randomDuration).Err(); err != nil {
					hlog.Error("redis.FavoriteAction: 设置redis缓存过期时间失败, err: ", err)
					return nil, err
				}
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
	}

	// 先查询作者的ID
	var authorID int64
	if err := db.Model(&models.Video{}).Where("id = ?", videoID).Select("author_id").Find(&authorID).Error; err != nil {
		hlog.Error("mysql.FavoriteAction: 查询作者的ID失败, err: ", err)
		return err
	}

	// 保存用户点赞视频的记录, 采用延迟双删策略
	// 删除redis缓存
	if err := rdb.SRem(context.Background(), key, userID).Err(); err != nil {
		hlog.Error("redis.FavoriteAction: 删除redis缓存失败, err: ", err)
	}

	// 更新favorite表
	if actionType == 1 {
		if err := db.Create(&models.Favorite{UserID: userID, VideoID: videoID}).Error; err != nil {
			hlog.Error("mysql.FavoriteAction: 更新favorite表失败, err: ", err)
			return err
		}
	} else {
		if err := db.Where("user_id = ? AND video_id = ?", userID, videoID).Delete(&models.Favorite{}).Error; err != nil {
			hlog.Error("mysql.FavoriteAction: 更新favorite表失败, err: ", err)
			return err
		}
	}

	// 延迟后删除redis缓存
	go func() {
		time.Sleep(delayTime)
		if err := rdb.SRem(context.Background(), key, userID).Err(); err != nil {
			hlog.Error("redis.FavoriteAction: 删除redis缓存失败, err: ", err)
		}
	}()

	// 更新video的favorite_count字段
	if err := rdb.IncrBy(context.Background(), getRedisKey(KeyVideoFavoriteCountPF+strconv.FormatInt(videoID, 10)), actionType).Err(); err != nil {
		hlog.Error("redis.FavoriteAction: 更新video的favorite_count字段失败, err: ", err)
		return err
	}

	// 更新user当前用户的favorite_count字段
	if err := rdb.IncrBy(context.Background(), getRedisKey(KeyUserFavoriteCountPF+strconv.FormatInt(userID, 10)), actionType).Err(); err != nil {
		hlog.Error("redis.FavoriteAction: 更新user当前用户的favorite_count字段失败, err: ", err)
		return err
	}

	// 更新user作者的total_favorited字段
	if err := rdb.IncrBy(context.Background(), getRedisKey(KeyUserTotalFavoritedPF+strconv.FormatInt(authorID, 10)), actionType).Err(); err != nil {
		hlog.Error("redis.FavoriteAction: 更新user作者的total_favorited字段失败, err: ", err)
		return err
	}

	// 写入待同步切片
	lock.Lock()
	cacheUserID = append(cacheUserID, userID, authorID)
	cacheVideoIDs = append(cacheVideoIDs, videoID)
	lock.Unlock()

	return nil
}

func GetFavoriteList(userID int64) ([]int64, error) {
	// 先查询redis缓存
	key := getRedisKey(KeyUserFavoritePF + strconv.FormatInt(userID, 10))
	videoIDStrs := rdb.SMembers(context.Background(), key).Val()
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
		if err := db.Model(&models.Favorite{}).Where("user_id = ?", userID).Select("video_id").Find(&videoIDs).Error; err != nil {
			hlog.Error("mysql.GetFavoriteList: 查询favorite表失败, err: ", err)
			return nil, err
		}

		// 写入redis缓存
		if len(videoIDs) > 0 {
			go func() {
				pipeline := rdb.Pipeline()
				for _, videoID := range videoIDs {
					pipeline.SAdd(context.Background(), key, videoID)
				}
				pipeline.Expire(context.Background(), key, expireTime+randomDuration)
				if _, err := pipeline.Exec(context.Background()); err != nil {
					hlog.Error("redis.GetFavoriteList: 写入redis缓存失败, err: ", err)
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
