package dal

import (
	"context"
	"strconv"
	"time"

	"douyin/src/dal/model"
	"douyin/src/pkg/snowflake"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func CheckFavoriteExist(ctx context.Context, userID int64, videoID int64) (bool, error) {
	// 查看是否已经点赞
	key := GetRedisKey(KeyUserFavoritePF, strconv.FormatInt(userID, 10))
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
		RDB.SAdd(ctx, key, videoID)
		RDB.Expire(ctx, key, 0)

		return true, nil
	}

	return false, nil
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
	key := GetRedisKey(KeyUserFavoritePF, strconv.FormatInt(userID, 10))
	_, err, _ = g.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(delayTime)
			g.Forget(key)
		}()

		// 先查询redis缓存
		videoIDStrs, err := RDB.SMembers(ctx, key).Result()
		if err == redis.Nil {
			// 缓存未命中，查询mysql
			if err := qFavorite.WithContext(ctx).Where(qFavorite.UserID.Eq(userID)).Select(qFavorite.VideoID).Scan(&videoIDs); err != nil {
				return nil, err
			}

			// 写入redis缓存
			if len(videoIDs) > 0 {
				RDB.SAdd(ctx, key, videoIDs)
				RDB.Expire(ctx, key, 0)
			}

			return nil, err
		} else if err != nil {
			return nil, err
		} else {
			// 缓存命中，转换为int64
			videoIDs = make([]int64, len(videoIDStrs))
			for i, videoIDStr := range videoIDStrs {
				videoID, err := strconv.ParseInt(videoIDStr, 10, 64)
				if err != nil {
					return nil, err
				}
				videoIDs[i] = videoID
			}

			return nil, nil
		}
	})

	return
}

// GetUserFavoriteCount 获取用户点赞数
func GetUserFavoriteCount(ctx context.Context, userID int64) (cnt int64, err error) {
	// 使用singleflight解决缓存击穿并减少redis压力
	key := GetRedisKey(KeyUserFavoriteCountPF, strconv.FormatInt(userID, 10))
	_, err, _ = g.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(delayTime)
			g.Forget(key)
		}()

		// 先查询redis缓存
		cnt, err = RDB.Get(ctx, key).Int64()
		if err == redis.Nil {
			// 缓存未命中，查询mysql
			cnt, err = qFavorite.WithContext(ctx).Where(qFavorite.UserID.Eq(userID)).Count()
			if err != nil {
				return nil, err
			}

			// 写入redis缓存
			err = RDB.Set(ctx, key, cnt, 0).Err()
			return nil, err
		}

		return nil, err
	})

	return
}

func GetVideoFavoriteCount(ctx context.Context, videoID int64) (cnt int64, err error) {
	// 使用singleflight解决缓存击穿并减少redis压力
	key := GetRedisKey(KeyVideoFavoriteCountPF, strconv.FormatInt(videoID, 10))
	_, err, _ = g.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(delayTime)
			g.Forget(key)
		}()

		// 先查询redis缓存
		cnt, err = RDB.Get(ctx, key).Int64()
		if err == redis.Nil {
			// 缓存未命中，查询mysql
			cnt, err = qFavorite.WithContext(ctx).Where(qFavorite.VideoID.Eq(videoID)).Count()
			if err != nil {
				return nil, err
			}

			// 写入redis缓存
			err = RDB.Set(ctx, key, cnt, 0).Err()
			return nil, err
		}

		return nil, err
	})

	return
}
