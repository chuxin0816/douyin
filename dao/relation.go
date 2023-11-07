package dao

import (
	"context"
	"douyin/models"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func RelationAction(userID, toUserID int64, actionType int64) error {
	// 查看是否关注
	key := getRedisKey(KeyUserFollowerPF + strconv.FormatInt(toUserID, 10))
	exist := rdb.SIsMember(context.Background(), key, userID).Val()
	if exist && actionType == 1 {
		hlog.Error("mysql.RelationAction 已经关注过了")
		return ErrAlreadyFollow
	}

	// 缓存未命中, 查询数据库
	if !exist {
		// 使用singleflight避免缓存击穿
		_, err, _ := g.Do(key, func() (interface{}, error) {
			go func() {
				time.Sleep(delayTime)
				g.Forget(key)
			}()
			relation := &models.Relation{}
			if err := db.Where("user_id = ? AND follower_id = ?", toUserID, userID).Find(relation).Error; err != nil {
				hlog.Error("mysql.RelationAction 查看是否关注失败, err: ", err)
				return nil, err
			}
			if relation.ID != 0 && actionType == 1 {
				// 写入redis缓存
				go func() {
					if err := rdb.SAdd(context.Background(), key, userID).Err(); err != nil {
						hlog.Error("redis.RelationAction 写入redis缓存失败, err: ", err)
						return
					}
					if err := rdb.Expire(context.Background(), key, expireTime+randomDuration).Err(); err != nil {
						hlog.Error("redis.RelationAction 设置redis缓存过期时间失败, err: ", err)
						return
					}
				}()
				return nil, ErrAlreadyFollow
			}
			if relation.ID == 0 && actionType == -1 {
				return nil, ErrNotFollow
			}
			return nil, nil
		})
		if err != nil {
			return err
		}
	}

	// 更新relation表
	if actionType == 1 {
		if err := db.Create(&models.Relation{UserID: toUserID, FollowerID: userID}).Error; err != nil {
			hlog.Error("mysql.RelationAction 更新relation表失败, err: ", err)
		}
	} else {
		if err := db.Where("user_id = ? AND follower_id = ?", toUserID, userID).Delete(&models.Relation{}).Error; err != nil {
			hlog.Error("mysql.RelationAction 更新relation表失败, err: ", err)
		}
	}

	// 更新user的follow_count和follower_count字段
	if err := rdb.IncrBy(context.Background(), getRedisKey(KeyUserFollowCountPF+strconv.FormatInt(userID, 10)), actionType).Err(); err != nil {
		hlog.Error("redis.RelationAction 更新user的follow_count字段失败, err: ", err)
		return err
	}
	if err := rdb.IncrBy(context.Background(), getRedisKey(KeyUserFollowerCountPF+strconv.FormatInt(toUserID, 10)), actionType).Err(); err != nil {
		hlog.Error("redis.RelationAction 更新user的follower_count字段失败, err: ", err)
	}

	return nil
}

func FollowList(toUserID int64) ([]*models.User, error) {
	// 查询用户ID列表
	var userIDList []int64
	err := db.Table("relations").Select("user_id").Where("follower_id = ?", toUserID).Find(&userIDList).Error
	if err != nil {
		hlog.Error("mysql.FollowList 查询用户ID列表失败, err: ", err)
		return nil, err
	}

	// 查询用户列表
	userList, err := GetUserByIDs(userIDList)
	if err != nil {
		hlog.Error("mysql.FollowList 查询用户列表失败, err: ", err)
		return nil, err
	}

	return userList, nil
}

func FollowerList(toUserID int64) ([]*models.User, error) {
	// 查询粉丝ID列表
	var followerIDList []int64
	err := db.Table("relations").Select("follower_id").Where("user_id = ?", toUserID).Find(&followerIDList).Error
	if err != nil {
		hlog.Error("mysql.FollowerList 查询粉丝ID列表失败, err: ", err)
		return nil, err
	}

	// 查询粉丝列表
	followerList, err := GetUserByIDs(followerIDList)
	if err != nil {
		hlog.Error("mysql.FollowerList 查询粉丝列表失败, err: ", err)
		return nil, err
	}

	return followerList, nil
}
