package dal

import (
	"context"
	"strconv"
	"time"

	"douyin/dal/model"
	"douyin/pkg/snowflake"

	"gorm.io/gorm"
)

func CheckFavoriteExist(ctx context.Context, userID int64, videoID int64) (bool, error) {
	// 查看是否已经点赞
	key := GetRedisKey(KeyUserFavoritePF + strconv.FormatInt(userID, 10))
	// 使用singleflight避免缓存击穿和减少缓存压力
	exist, err, _ := g.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(delayTime)
			g.Forget(key)
		}()
		exist := RDB.SIsMember(ctx, key, videoID).Val()
		if exist {
			return true, nil
		}
		// 缓存未命中，查询mysql中是否有记录
		var id int64
		if err := qFavorite.WithContext(ctx).Where(qFavorite.UserID.Eq(userID), qFavorite.VideoID.Eq(videoID)).Select(qFavorite.ID).Scan(&id); err != nil {
			if err == gorm.ErrRecordNotFound {
				return false, nil
			}
			return false, err
		}
		if id != 0 {
			// 写入redis缓存
			go func() {
				RDB.SAdd(ctx, key, videoID)
				RDB.Expire(ctx, key, ExpireTime+GetRandomTime())
			}()
			return true, nil
		}

		return false, nil
	})

	return exist.(bool), err
}

func CreateFavorite(ctx context.Context, userID, videoID int64) error {
	mFavorite := &model.Favorite{
		ID:      snowflake.GenerateID(),
		UserID:  userID,
		VideoID: videoID,
	}
	return qFavorite.WithContext(ctx).Create(mFavorite)
}

func DeleteFavorite(ctx context.Context, userID, videoID int64) error {
	_, err := qFavorite.WithContext(ctx).Where(qFavorite.UserID.Eq(userID), qFavorite.VideoID.Eq(videoID)).Delete()
	return err
}

func GetFavoriteList(ctx context.Context, userID int64) (videoIDs []int64, err error) {
	// 使用singleflight防止缓存击穿并减少redis压力
	key := GetRedisKey(KeyUserFavoritePF + strconv.FormatInt(userID, 10))
	_, err, _ = g.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(delayTime)
			g.Forget(key)
		}()

		// 先查询redis缓存
		videoIDStrs := RDB.SMembers(ctx, key).Val()
		if len(videoIDStrs) != 0 {
			videoIDs = make([]int64, len(videoIDStrs))
			for i, videoIDStr := range videoIDStrs {
				videoID, _ := strconv.ParseInt(videoIDStr, 10, 64)
				videoIDs[i] = videoID
			}
			return nil, nil
		}

		// 缓存未命中，查询mysql
		if err := qFavorite.WithContext(ctx).Where(qFavorite.UserID.Eq(userID)).Select(qFavorite.VideoID).Scan(&videoIDs); err != nil {
			return nil, err
		}

		// 写入redis缓存
		if len(videoIDs) > 0 {
			go func() {
				pipeline := RDB.Pipeline()
				for _, videoID := range videoIDs {
					pipeline.SAdd(ctx, key, videoID)
				}
				pipeline.Expire(ctx, key, ExpireTime+GetRandomTime())
				_, _ = pipeline.Exec(ctx)
			}()
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return videoIDs, nil
}
