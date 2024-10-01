package dal

import (
	"context"
	"strconv"
	"time"

	"douyin/src/common/snowflake"
	"douyin/src/dal/model"

	"github.com/allegro/bigcache/v3"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func CreateComment(ctx context.Context, comment *model.Comment) error {
	comment.ID = snowflake.GenerateID()
	return qComment.WithContext(ctx).Create(comment)
}

func DeleteComment(ctx context.Context, commentID int64) (err error) {
	_, err = qComment.WithContext(ctx).Where(qComment.ID.Eq(commentID)).Delete()
	return
}

func GetCommentByID(ctx context.Context, commentID int64) (*model.Comment, error) {
	comment, err := qComment.WithContext(ctx).
		Where(qComment.ID.Eq(commentID)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCommentNotExist
		}
		return nil, err
	}
	return comment, nil
}

func GetCommentList(ctx context.Context, videoID int64) ([]*model.Comment, error) {
	commentList, err := qComment.WithContext(ctx).Where(qComment.VideoID.Eq(videoID)).Limit(30).Find()
	if err != nil {
		return nil, err
	}
	return commentList, nil
}

// GetVideoFavoriteCount 获取视频点赞数
func GetVideoCommentCount(ctx context.Context, videoID int64) (count int64, err error) {
	key := GetRedisKey(KeyVideoCommentCountPF, strconv.FormatInt(videoID, 10))
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
		count, err = RDB.Get(ctx, key).Int64()
		if err == redis.Nil {
			// 缓存未命中，查询mysql
			count, err = qComment.WithContext(ctx).Where(qComment.VideoID.Eq(videoID)).Count()
			if err != nil {
				return nil, err
			}

			// 写入redis缓存
			err = RDB.Set(ctx, key, count, ExpireTime+GetRandomTime()).Err()
		} else if err == nil {
			// 写入本地缓存
			if err := Cache.Set(key, []byte(strconv.FormatInt(count, 10))); err != nil {
				klog.Error("Cache.Set failed, err: ", err)
			}
		}
		return nil, err
	})

	return
}
