package dal

import (
	"context"
	"douyin/dal/model"

	"strconv"
	"time"

	"gorm.io/gorm"
)

func CheckRelationExist(ctx context.Context, userID, toUserID int64) (bool, error) {
	key := GetRedisKey(KeyUserFollowerPF + strconv.FormatInt(toUserID, 10))
	// 使用singleflight避免缓存击穿和减少缓存压力
	exist, err, _ := g.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(delayTime)
			g.Forget(key)
		}()
		if RDB.SIsMember(ctx, key, userID).Val() {
			return true, nil
		}

		// 缓存未命中, 查询数据库
		relation, err := qRelation.WithContext(ctx).Where(qRelation.UserID.Eq(toUserID), qRelation.FollowerID.Eq(userID)).
			Select(qRelation.ID).First()
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return false, nil
			}
			return false, err
		}

		if relation.ID != 0 {
			// 写入redis缓存
			go func() {
				RDB.SAdd(ctx, key, userID)
				RDB.Expire(ctx, key, ExpireTime+GetRandomTime())
			}()
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return false, err
	}

	return exist.(bool), nil
}

func Follow(ctx context.Context, userID, toUserID int64) error {
	return qRelation.WithContext(ctx).Create(&model.Relation{UserID: toUserID, FollowerID: userID})
}

func UnFollow(ctx context.Context, userID, toUserID int64) error {
	_, err := qRelation.WithContext(ctx).Where(qRelation.UserID.Eq(toUserID), qRelation.FollowerID.Eq(userID)).Delete()
	return err
}
func FollowList(ctx context.Context, toUserID int64) ([]*model.User, error) {
	// 查询用户ID列表
	var userIDList []int64
	if err := qRelation.WithContext(ctx).Where(qRelation.FollowerID.Eq(toUserID)).Select(qRelation.UserID).Scan(&userIDList); err != nil {
		return nil, err
	}

	// 查询用户列表
	userList, err := GetUserByIDs(ctx, userIDList)
	if err != nil {
		return nil, err
	}

	return userList, nil
}

func FollowerList(ctx context.Context, toUserID int64) ([]*model.User, error) {
	// 查询粉丝ID列表
	var followerIDList []int64
	if err := qRelation.WithContext(ctx).Where(qRelation.UserID.Eq(toUserID)).Select(qRelation.FollowerID).Scan(&followerIDList); err != nil {
		return nil, err
	}

	// 查询粉丝列表
	followerList, err := GetUserByIDs(ctx, followerIDList)
	if err != nil {
		return nil, err
	}

	return followerList, nil
}
