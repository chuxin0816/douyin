package dao

import (
	"context"
	"douyin/models"
	"errors"
	"math/rand"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/gorm"
)

const (
	expireTime = time.Hour * 72
	timeout    = time.Second * 5
	randFactor = 30
	lock       = "favoriteLock"
)

var (
	ErrAlreadyFavorite = errors.New("已经点赞过了")
	ErrNotFavorite     = errors.New("还没有点赞过")
	randomDuration     = time.Duration(rand.Intn(randFactor)) * time.Minute
)

func FavoriteAction(userID int64, videoID int64, actionType int64) error {
	videoIDStr := strconv.FormatInt(videoID, 10)
	key := getRedisKey(KeyVideoLikerPF) + videoIDStr

	// 查看是否已经点赞
	exist := rdb.SIsMember(context.Background(), key, userID).Val()
	if exist && actionType == 1 {
		hlog.Error("redis.FavoriteAction: 已经点赞过了")
		return ErrAlreadyFavorite
	}
	if !exist {
		// 缓存未命中，查询mysql中是否有记录
		var id int64
		if err := db.Model(&models.Favorite{}).Where("user_id = ? AND video_id = ?", userID, videoID).Select("id").Scan(&id).Error; err != nil {
			hlog.Error("mysql.FavoriteAction: 查询mysql中是否有记录失败, err: ", err)
			return err
		}
		// mysql中有记录
		if id != 0 && actionType == 1 {
			// 写入redis缓存
			go func() {
				ok, err := rdb.SetNX(context.Background(), lock, 1, timeout).Result()
				if err != nil {
					hlog.Error("redis.FavoriteAction: 加锁失败, err: ", err)
					return
				}
				if !ok {
					return
				}
				if err := rdb.SAdd(context.Background(), key, userID).Err(); err != nil {
					hlog.Error("redis.FavoriteAction: 写入redis缓存失败, err: ", err)
					return
				}
				if err := rdb.Expire(context.Background(), key, expireTime+randomDuration).Err(); err != nil {
					hlog.Error("redis.FavoriteAction: 设置redis缓存过期时间失败, err: ", err)
					return
				}
				if err := rdb.Del(context.Background(), lock).Err(); err != nil {
					hlog.Error("redis.FavoriteAction: 释放锁失败, err: ", err)
					return
				}
			}()
			hlog.Error("redis.FavoriteAction: 已经点赞过了")
			return ErrAlreadyFavorite
		}
		// mysql中没有记录
		if id == 0 && actionType == -1 {
			hlog.Error("redis.FavoriteAction: 还没有点赞过")
			return ErrNotFavorite
		}
	}

	// 先查询作者的ID
	var authorID int64
	err := db.Model(&models.Video{}).Where("id = ?", videoID).Select("author_id").Scan(&authorID).Error
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

	// 删除redis缓存
	if err := rdb.SRem(context.Background(), key, userID).Err(); err != nil {
		hlog.Error("redis.FavoriteAction: 删除redis缓存失败, err: ", err)
		return err
	}

	return nil
}

func GetFavoriteList(userID int64) ([]int64, error) {
	var favoriteList []*models.Favorite
	err := db.Where("user_id = ?", userID).Find(&favoriteList).Error
	if err != nil {
		hlog.Error("mysql.GetFavoriteList: 查询favorite表失败, err: ", err)
		return nil, err
	}

	// 将models.Favorite视频ID提取出来
	videoIDs := make([]int64, 0, len(favoriteList))
	for _, favorite := range favoriteList {
		videoIDs = append(videoIDs, favorite.VideoID)
	}
	return videoIDs, nil
}
