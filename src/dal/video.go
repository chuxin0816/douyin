package dal

import (
	"context"
	"strconv"
	"sync/atomic"
	"time"

	"douyin/src/dal/model"
	"douyin/src/pkg/snowflake"

	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
	"gorm.io/gorm"
)

const (
	videoPrefix = "http://oss.chuxin0816.com/video/"
	imagePrefix = "http://oss.chuxin0816.com/image/"
)

// GetVideoByID 通过视频ID查询视频信息
func GetVideoByID(ctx context.Context, videoID int64) (video *model.Video, err error) {
	// 使用singleflight解决缓存击穿并减少redis压力
	key := GetRedisKey(KeyVideoInfoPF, strconv.FormatInt(videoID, 10))
	_, err, _ = g.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(delayTime)
			g.Forget(key)
		}()

		// 先查询redis缓存
		videoInfo, err := RDB.Get(ctx, key).Bytes()
		if err == redis.Nil {
			// 缓存未命中，查询mysql
			video, err = qVideo.WithContext(ctx).Where(qVideo.ID.Eq(videoID)).First()
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					return nil, ErrVideoNotExist
				}
				return nil, err
			}

			// 写入redis缓存
			videoInfo, err := msgpack.Marshal(video)
			if err != nil {
				return nil, err
			}
			err = RDB.Set(ctx, key, videoInfo, 0).Err()

			return nil, err
		} else if err != nil {
			return nil, err
		} else {
			// 缓存命中
			err = msgpack.Unmarshal(videoInfo, video)
			return nil, err
		}
	})

	return
}

// GetFeedList 获取视频Feed流
func GetFeedList(ctx context.Context, userID *int64, latestTime time.Time, count int) ([]*model.Video, error) {
	// 查询Feed流ID列表
	var feedIDs []int64
	if err := qVideo.WithContext(ctx).Where(qVideo.UploadTime.Lt(latestTime)).Select(qVideo.ID).Limit(count).Scan(&feedIDs); err != nil {
		return nil, err
	}

	videoList := make([]*model.Video, len(feedIDs))
	// 查询视频信息
	for i, videoID := range feedIDs {
		video, err := GetVideoByID(ctx, videoID)
		if err != nil {
			return nil, err
		}
		videoList[i] = video
	}

	return videoList, nil
}

// SaveVideo 保存视频信息到数据库
func SaveVideo(ctx context.Context, userID int64, videoName, coverName, title string) error {
	video := &model.Video{
		ID:         snowflake.GenerateID(),
		AuthorID:   userID,
		PlayURL:    videoPrefix + videoName,
		CoverURL:   imagePrefix + coverName,
		UploadTime: time.Now(),
		Title:      title,
	}
	// 添加到布隆过滤器
	bloomFilter.Add([]byte(strconv.FormatInt(video.ID, 10)))

	return qVideo.WithContext(ctx).Create(video)
}

// GetUserTotalFavorited 获取用户发布的视频ID列表
func GetUserTotalFavorited(ctx context.Context, userID int64) (total int64, err error) {
	// 使用singleflight解决缓存击穿并减少redis压力
	key := GetRedisKey(KeyUserTotalFavoritedPF, strconv.FormatInt(userID, 10))
	_, err, _ = g.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(delayTime)
			g.Forget(key)
		}()

		// 先查询redis缓存
		total, err = RDB.Get(ctx, key).Int64()
		if err == redis.Nil {
			// 缓存未命中，查询mysql
			// 查询用户发布列表
			var videoIDs []int64
			if err := qVideo.WithContext(ctx).Where(qVideo.AuthorID.Eq(userID)).Select(qVideo.ID).Scan(&videoIDs); err != nil {
				return nil, err
			}

			// 查询用户发布视频的点赞数
			for _, videoID := range videoIDs {
				cnt, err := GetVideoFavoriteCount(ctx, videoID)
				if err != nil {
					return nil, err
				}
				atomic.AddInt64(&total, cnt)
			}

			// 写入redis缓存
			err = RDB.Set(ctx, key, total, 0).Err()
			return nil, err
		}
		return nil, err
	})

	return
}

// GetPublishList 获取用户发布的视频列表
func GetPublishList(ctx context.Context, authorID int64) ([]*model.Video, error) {
	// 查询视频ID列表
	var videoIDs []int64
	err := qVideo.WithContext(ctx).Where(qVideo.AuthorID.Eq(authorID)).Select(qVideo.ID).Scan(&videoIDs)
	if err != nil {
		return nil, err
	}

	// 查询视频信息
	return GetVideoList(ctx, videoIDs)
}

func GetVideoList(ctx context.Context, videoIDs []int64) ([]*model.Video, error) {
	videoList := make([]*model.Video, len(videoIDs))
	for i, videoID := range videoIDs {
		video, err := GetVideoByID(ctx, videoID)
		if err != nil {
			return nil, err
		}
		videoList[i] = video
	}

	return videoList, nil
}

func GetUserWorkCount(ctx context.Context, userID int64) (cnt int64, err error) {
	// 使用singleflight解决缓存击穿并减少redis压力
	key := GetRedisKey(KeyUserWorkCountPF, strconv.FormatInt(userID, 10))
	_, err, _ = g.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(delayTime)
			g.Forget(key)
		}()

		// 先查询redis缓存
		cnt, err = RDB.Get(ctx, key).Int64()
		if err == redis.Nil {
			// 缓存未命中，查询mysql
			cnt, err = qVideo.WithContext(ctx).Where(qVideo.AuthorID.Eq(userID)).Count()
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

func CheckVideoExist(ctx context.Context, videoID int64) error {
	// 判断视频是否存在
	if !bloomFilter.Test([]byte(strconv.FormatInt(videoID, 10))) {
		return ErrVideoNotExist
	}

	_, err := qVideo.WithContext(ctx).
		Where(qVideo.ID.Eq(videoID)).
		Select(qVideo.ID).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrVideoNotExist
		}
		return err
	}

	return nil
}

func GetAuthorID(ctx context.Context, videoID int64) (int64, error) {
	// 先查询作者的ID
	var authorID int64
	err := qVideo.WithContext(ctx).Where(qVideo.ID.Eq(videoID)).Select(qVideo.AuthorID).Scan(&authorID)

	return authorID, err
}
