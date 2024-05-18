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
	// 使用singleflight解决缓存击穿并减少redis压力
	key := GetRedisKey(KeyVideoCommentCountPF, strconv.FormatInt(videoID, 10))
	_, err, _ = g.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(delayTime)
			g.Forget(key)
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
			err = RDB.Set(ctx, key, count, 0).Err()
			return nil, err
		}

		return nil, err
	})

	return
}
