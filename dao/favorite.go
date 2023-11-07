package dao

import (
	"context"
	"douyin/models"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/gorm"
)

func FavoriteAction(userID int64, videoID int64, actionType int64) error {
	// 先判断布隆过滤器中是否存在
	if !bloomFilter.Test([]byte(strconv.FormatInt(videoID, 10))) {
		hlog.Error("mysql.FavoriteAction: 视频不存在, videoID: ", videoID)
		return ErrVideoNotExist
	}
	videoIDStr := strconv.FormatInt(videoID, 10)
	key := getRedisKey(KeyVideoFavoritePF) + videoIDStr

	// 查看是否已经点赞
	exist := rdb.SIsMember(context.Background(), key, userID).Val()
	if exist && actionType == 1 {
		hlog.Error("redis.FavoriteAction: 已经点赞过了")
		return ErrAlreadyFavorite
	}
	if !exist {
		// 缓存未命中，查询mysql中是否有记录
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
				hlog.Error("redis.FavoriteAction: 已经点赞过了")
				return nil, ErrAlreadyFavorite
			}
			// mysql中没有记录
			if id == 0 && actionType == -1 {
				hlog.Error("redis.FavoriteAction: 还没有点赞过")
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
	err := db.Model(&models.Video{}).Where("id = ?", videoID).Select("author_id").Find(&authorID).Error
	if err != nil {
		hlog.Error("mysql.FavoriteAction: 查询作者的ID失败, err: ", err)
		return err
	}

	// 保存用户点赞视频的记录, 采用延迟双删策略
	// 删除redis缓存
	if err := rdb.SRem(context.Background(), key, userID).Err(); err != nil {
		hlog.Error("redis.FavoriteAction: 删除redis缓存失败, err: ", err)
		return err
	}

	// 开启事务
	tx := db.Begin()

	// 更新favorite表
	if actionType == 1 {
		err = tx.Create(&models.Favorite{UserID: userID, VideoID: videoID}).Error
	} else {
		err = tx.Where("user_id = ? AND video_id = ?", userID, videoID).Delete(&models.Favorite{}).Error
	}
	if err != nil {
		tx.Rollback()
		hlog.Error("mysql.FavoriteAction: 更新favorite表失败, err: ", err)
		return err
	}

	// 更新video表中的favorite_count字段
	err = tx.Model(&models.Video{}).Where("id = ?", videoID).Update("favorite_count", gorm.Expr("favorite_count + ?", actionType)).Error
	if err != nil {
		tx.Rollback()
		hlog.Error("mysql.FavoriteAction: 更新video表中的favorite_count字段失败, err: ", err)
		return err
	}

	// 更新user表中当前用户的favorite_count字段
	err = tx.Model(models.User{}).Where("id = ?", userID).Update("favorite_count", gorm.Expr("favorite_count + ?", actionType)).Error
	if err != nil {
		tx.Rollback()
		hlog.Error("mysql.FavoriteAction: 更新user表中当前用户的favorite_count字段失败, err: ", err)
		return err
	}

	// 更新user表中作者的total_favorited字段
	err = tx.Model(models.User{}).Where("id = ?", authorID).Update("total_favorited", gorm.Expr("total_favorited + ?", actionType)).Error
	if err != nil {
		tx.Rollback()
		hlog.Error("mysql.FavoriteAction: 更新user表中作者的total_favorited字段失败, err: ", err)
		return err
	}

	// 提交事务
	tx.Commit()

	// 延迟后删除redis缓存
	time.Sleep(delayTime)
	if err := rdb.SRem(context.Background(), key, userID).Err(); err != nil {
		hlog.Error("redis.FavoriteAction: 删除redis缓存失败, err: ", err)
		return err
	}

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

	// 查询mysql
	if err := db.Model(&models.Favorite{}).Where("user_id = ?", userID).Select("video_id").Find(&videoIDs).Error; err != nil {
		hlog.Error("mysql.GetFavoriteList: 查询favorite表失败, err: ", err)
		return nil, err
	}

	// 写入redis缓存
	go func() {
		pipeline := rdb.Pipeline()
		for _, videoID := range videoIDs {
			pipeline.SAdd(context.Background(), key, videoID)
		}
		pipeline.Expire(context.Background(), key, expireTime+randomDuration)
		_, err := pipeline.Exec(context.Background())
		if err != nil {
			hlog.Error("redis.GetFavoriteList: 写入redis缓存失败, err: ", err)
		}
	}()

	return videoIDs, nil
}