package dal

import (
	"context"
	"strconv"
	"time"

	"douyin/dal/model"
	"douyin/pkg/snowflake"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// CheckRelationExist 检查userID是否关注了toUserID
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
	relation := &model.Relation{
		ID:         snowflake.GenerateID(),
		UserID:     toUserID,
		FollowerID: userID,
	}
	return qRelation.WithContext(ctx).Create(relation)
}

func UnFollow(ctx context.Context, userID, toUserID int64) error {
	_, err := qRelation.WithContext(ctx).Where(qRelation.UserID.Eq(toUserID), qRelation.FollowerID.Eq(userID)).Delete()
	return err
}

func FollowList(ctx context.Context, userID int64) (followList []int64, err error) {
	// 使用singleflight防止缓存击穿并减少redis压力
	key := GetRedisKey(KeyUserFollowPF + strconv.FormatInt(userID, 10))
	_, err, _ = g.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(delayTime)
			g.Forget(key)
		}()

		// 先查询redis缓存
		userIDs, err := RDB.SMembers(ctx, key).Result()
		if err == redis.Nil {
			// 缓存未命中, 查询数据库
			if err := qRelation.WithContext(ctx).Where(qRelation.FollowerID.Eq(userID)).Select(qRelation.UserID).Scan(&followList); err != nil {
				return nil, err
			}

			// 写入redis缓存
			if len(followList) > 0 {
				go func() {
					pipe := RDB.Pipeline()
					pipe.SAdd(ctx, key, followList)
					pipe.Expire(ctx, key, ExpireTime+GetRandomTime())
					_, _ = pipe.Exec(ctx)
				}()
			}

			return nil, nil
		} else if err != nil {
			return nil, err
		} else {
			for _, id := range userIDs {
				userID, err := strconv.ParseInt(id, 10, 64)
				if err != nil {
					return nil, err
				}
				followList = append(followList, userID)
			}

			return nil, nil
		}
	})

	return
}

func FollowerList(ctx context.Context, userID int64) (followerList []int64, err error) {
	// 使用singleflight防止缓存击穿并减少redis压力
	key := GetRedisKey(KeyUserFollowerPF + strconv.FormatInt(userID, 10))
	_, err, _ = g.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(delayTime)
			g.Forget(key)
		}()
		// 先查询redis缓存
		userIDs, err := RDB.SMembers(ctx, key).Result()
		if err == redis.Nil {
			// 缓存未命中，查询mysql
			if err := qRelation.WithContext(ctx).Where(qRelation.UserID.Eq(userID)).Select(qRelation.FollowerID).Scan(&followerList); err != nil {
				return nil, err
			}

			// 写入redis缓存
			if len(followerList) > 0 {
				go func() {
					pipeline := RDB.Pipeline()
					pipeline.SAdd(ctx, key, followerList)
					pipeline.Expire(ctx, key, ExpireTime+GetRandomTime())
					_, _ = pipeline.Exec(ctx)
				}()
			}
			return nil, nil
		} else if err != nil {
			return nil, err
		} else {
			// 缓存命中，转换为int64
			followerList = make([]int64, len(userIDs))
			for i, id := range userIDs {
				userID, err := strconv.ParseInt(id, 10, 64)
				if err != nil {
					return nil, err
				}
				followerList[i] = userID
			}
			return nil, nil
		}
	})

	return
}
