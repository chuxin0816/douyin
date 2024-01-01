package dal

import (
	"context"
	"douyin/dal/model"

	"strconv"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

func RelationAction(userID, toUserID int64, actionType int64) error {
	// 查看是否关注
	key := getRedisKey(KeyUserFollowerPF + strconv.FormatInt(toUserID, 10))
	// 使用singleflight避免缓存击穿和减少缓存压力
	_, err, _ := g.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(delayTime)
			g.Forget(key)
		}()
		exist := rdb.SIsMember(context.Background(), key, userID).Val()
		if exist {
			if actionType == 1 {
				klog.Error("已经关注过了")
				return nil, ErrAlreadyFollow
			}
			return nil, nil
		}
		// 缓存未命中, 查询数据库
		relation, err := qRelation.WithContext(context.Background()).Where(qRelation.UserID.Eq(toUserID), qRelation.FollowerID.Eq(userID)).
			Select(qRelation.ID).First()
		if err != nil {
			klog.Error("查看是否关注失败, err: ", err)
			return nil, err
		}
		if relation.ID != 0 && actionType == 1 {
			// 写入redis缓存
			go func() {
				if err := rdb.SAdd(context.Background(), key, userID).Err(); err != nil {
					klog.Error("写入redis缓存失败, err: ", err)
					return
				}
				if err := rdb.Expire(context.Background(), key, expireTime+getRandomTime()).Err(); err != nil {
					klog.Error("设置redis缓存过期时间失败, err: ", err)
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

	// 使用延迟双删策略
	if err := rdb.SRem(context.Background(), key, userID).Err(); err != nil {
		klog.Error("延迟双删策略失败, err: ", err)
	}

	// 更新relation表
	if actionType == 1 {
		if err := qRelation.WithContext(context.Background()).Create(&model.Relation{UserID: toUserID, FollowerID: userID}); err != nil {
			klog.Error("更新relation表失败, err: ", err)
		}
	} else {
		if _, err := qRelation.WithContext(context.Background()).Where(qRelation.UserID.Eq(toUserID), qRelation.FollowerID.Eq(userID)).Delete(); err != nil {
			klog.Error("更新relation表失败, err: ", err)
		}
	}
	// 延迟后删除redis缓存, 由kafka任务处理
	go func() {
		time.Sleep(delayTime)
		if err := rdb.SRem(context.Background(), key, userID).Err(); err != nil {
			klog.Error("删除redis缓存失败, err: ", err)
		}
	}()

	// 更新user的follow_count和follower_count字段
	if err := rdb.IncrBy(context.Background(), getRedisKey(KeyUserFollowCountPF+strconv.FormatInt(userID, 10)), actionType).Err(); err != nil {
		klog.Error("更新user的follow_count字段失败, err: ", err)
		return err
	}
	if err := rdb.IncrBy(context.Background(), getRedisKey(KeyUserFollowerCountPF+strconv.FormatInt(toUserID, 10)), actionType).Err(); err != nil {
		klog.Error("更新user的follower_count字段失败, err: ", err)
	}

	// 写入待同步切片
	cacheUserID.Store(userID, struct{}{})
	cacheUserID.Store(toUserID, struct{}{})

	return nil
}

func FollowList(toUserID int64) ([]*model.User, error) {
	// 查询用户ID列表
	var userIDList []int64
	if err := qRelation.WithContext(context.Background()).Where(qRelation.FollowerID.Eq(toUserID)).Select(qRelation.UserID).Scan(&userIDList); err != nil {
		klog.Error("查询用户ID列表失败, err: ", err)
		return nil, err
	}

	// 查询用户列表
	userList, err := GetUserByIDs(userIDList)
	if err != nil {
		klog.Error("查询用户列表失败, err: ", err)
		return nil, err
	}

	return userList, nil
}

func FollowerList(toUserID int64) ([]*model.User, error) {
	// 查询粉丝ID列表
	var followerIDList []int64
	if err := qRelation.WithContext(context.Background()).Where(qRelation.UserID.Eq(toUserID)).Select(qRelation.FollowerID).Scan(&followerIDList); err != nil {
		klog.Error("查询粉丝ID列表失败, err: ", err)
		return nil, err
	}

	// 查询粉丝列表
	followerList, err := GetUserByIDs(followerIDList)
	if err != nil {
		klog.Error("查询粉丝列表失败, err: ", err)
		return nil, err
	}

	return followerList, nil
}
