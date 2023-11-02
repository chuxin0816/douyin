package redis

import "context"

func FavoriteAction(userID int64, videoID int64, actionType int64) error {
	// 查看是否已经点赞
	rdb.S
	// 修改redis中的视频点赞数
	err := rdb.HIncrBy(context.Background(), getRedisKey(KeyVideoFavorite), videoID, int64(actionType)).Err()
}
