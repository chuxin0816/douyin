package redis

import (
	"context"
	"errors"
	"strconv"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

var (
	ErrAlreadyFavorite = errors.New("已经点赞过了")
	ErrNotFavorite     = errors.New("还没有点赞过")
)

func FavoriteAction(userID int64, videoID int64, actionType int64) error {
	videoIDStr := strconv.FormatInt(videoID, 10)
	// 查看是否已经点赞
	exist := rdb.SIsMember(context.Background(), getRedisKey(KeyVideoLikerPF)+videoIDStr, userID).Val()
	if exist && actionType == 1 {
		hlog.Error("redis.FavoriteAction: 已经点赞过了")
		return ErrAlreadyFavorite
	}
	if !exist && actionType == -1 {
		hlog.Error("redis.FavoriteAction: 还没有点赞过")
		return ErrNotFavorite
	}

	// 开启事务
	pipeline := rdb.TxPipeline()

	// 保存用户点赞视频的记录
	if actionType == 1 {
		pipeline.SAdd(context.Background(), getRedisKey(KeyVideoLikerPF)+videoIDStr, userID)
	} else {
		pipeline.SRem(context.Background(), getRedisKey(KeyVideoLikerPF)+videoIDStr, userID)
	}

	// 修改redis中的视频点赞数
	pipeline.HIncrBy(context.Background(), getRedisKey(KeyVideoFavorite), videoIDStr, int64(actionType)).Err()

	_, err := pipeline.Exec(context.Background())
	return err
}
