package dal

import (
	"context"
	"strconv"
	"sync/atomic"
	"time"

	"douyin/src/common/snowflake"
	"douyin/src/dal/model"

	"github.com/allegro/bigcache/v3"
	"github.com/cloudwego/kitex/pkg/klog"
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
	key := GetRedisKey(KeyVideoInfoPF, strconv.FormatInt(videoID, 10))
	// 查询本地缓存
	if val, err := Cache.Get(key); err == nil {
		err = msgpack.Unmarshal(val, video)
		if err == nil {
			return video, nil
		}
	} else if err != bigcache.ErrEntryNotFound {
		klog.Error("Cache.Get failed, err: ", err)
	}

	// 使用singleflight解决缓存击穿并减少redis压力
	_, err, _ = G.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(DelayTime)
			G.Forget(key)
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
			videoInfo, err = msgpack.Marshal(video)
			if err != nil {
				return nil, err
			}
			err = RDB.Set(ctx, key, videoInfo, ExpireTime+GetRandomTime()).Err()
		} else if err == nil {
			// 缓存命中
			err = msgpack.Unmarshal(videoInfo, video)
			if err != nil {
				return nil, err
			}
			// 写入本地缓存
			if err := Cache.Set(key, videoInfo); err != nil {
				klog.Error("Cache.Set failed, err: ", err)
			}
		}
		return nil, err
	})

	return
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

// GetFeedList 获取视频Feed流
func GetFeedList(ctx context.Context, userID *int64, latestTime time.Time, count int) ([]int64, error) {
	// 查询Feed流ID列表
	var feedIDs []int64
	if err := qVideo.WithContext(ctx).Where(qVideo.UploadTime.Lt(latestTime)).Select(qVideo.ID).Limit(count).Scan(&feedIDs); err != nil {
		return nil, err
	}

	return feedIDs, nil
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
	key := GetRedisKey(KeyUserTotalFavoritedPF, strconv.FormatInt(userID, 10))
	// 查询本地缓存
	if val, err := Cache.Get(key); err == nil {
		return strconv.ParseInt(string(val), 10, 64)
	} else if err != bigcache.ErrEntryNotFound {
		klog.Error("Cache.Get failed, err: ", err)
	}

	// 使用singleflight解决缓存击穿并减少redis压力
	_, err, _ = G.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(DelayTime)
			G.Forget(key)
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
			err = RDB.Set(ctx, key, total, ExpireTime+GetRandomTime()).Err()
		} else if err == nil {
			// 写入本地缓存
			if err := Cache.Set(key, []byte(strconv.FormatInt(total, 10))); err != nil {
				klog.Error("Cache.Set failed, err: ", err)
			}
		}
		return nil, err
	})

	return
}

// GetPublishList 获取用户发布的视频列表
func GetPublishList(ctx context.Context, authorID int64) ([]int64, error) {
	var videoIDs []int64
	err := qVideo.WithContext(ctx).Where(qVideo.AuthorID.Eq(authorID)).Select(qVideo.ID).Scan(&videoIDs)
	if err != nil {
		return nil, err
	}

	return videoIDs, nil
}

func GetUserWorkCount(ctx context.Context, userID int64) (cnt int64, err error) {
	key := GetRedisKey(KeyUserWorkCountPF, strconv.FormatInt(userID, 10))
	// 查询本地缓存
	if val, err := Cache.Get(key); err == nil {
		return strconv.ParseInt(string(val), 10, 64)
	} else if err != bigcache.ErrEntryNotFound {
		klog.Error("Cache.Get failed, err: ", err)
	}

	// 使用singleflight解决缓存击穿并减少redis压力
	_, err, _ = G.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(DelayTime)
			G.Forget(key)
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
			err = RDB.Set(ctx, key, cnt, ExpireTime+GetRandomTime()).Err()
		} else if err == nil {
			// 写入本地缓存
			if err := Cache.Set(key, []byte(strconv.FormatInt(cnt, 10))); err != nil {
				klog.Error("Cache.Set failed, err: ", err)
			}
		}
		return nil, err
	})

	return
}

func CheckVideoExist(ctx context.Context, videoID int64) (bool, error) {
	// 判断视频是否存在
	if !bloomFilter.Test([]byte(strconv.FormatInt(videoID, 10))) {
		return false, nil
	}

	_, err := qVideo.WithContext(ctx).Where(qVideo.ID.Eq(videoID)).Select(qVideo.ID).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func GetAuthorID(ctx context.Context, videoID int64) (int64, error) {
	// 先查询作者的ID
	var authorID int64
	err := qVideo.WithContext(ctx).Where(qVideo.ID.Eq(videoID)).Select(qVideo.AuthorID).Scan(&authorID)

	return authorID, err
}
