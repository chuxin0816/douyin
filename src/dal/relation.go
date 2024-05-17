package dal

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// CheckRelationExist 检查userID是否关注了authorID
func CheckRelationExist(ctx context.Context, userID, authorID int64) (bool, error) {
	key := GetRedisKey(KeyUserFollowPF, strconv.FormatInt(userID, 10))
	if RDB.SIsMember(ctx, key, authorID).Val() {
		return true, nil
	}

	// 缓存未命中, 查询数据库
	var builder strings.Builder
	builder.WriteString("match (v:user)-[:follow]->(v2:user) where id(v) == ")
	builder.WriteString(strconv.FormatInt(userID, 10))
	builder.WriteString(" and id(v2) == ")
	builder.WriteString(strconv.FormatInt(authorID, 10))
	builder.WriteString(" return count(*)>0 as result")
	resp, err := sessionPool.Execute(builder.String())
	if err != nil {
		return false, err
	}

	res, err := resp.GetValuesByColName("result")
	if err != nil {
		return false, err
	}
	exist, _ := res[0].AsBool()

	if exist {
		// 写入redis缓存
		RDB.SAdd(ctx, key, userID)
		RDB.Expire(ctx, key, ExpireTime+GetRandomTime())

		return true, nil
	}

	return false, nil
}

func Follow(ctx context.Context, userID, authorID int64) error {
	// 添加用户点
	var builder strings.Builder
	builder.WriteString("insert vertex user() values ")
	builder.WriteString(strconv.FormatInt(userID, 10))
	builder.WriteString(":()")
	_, err := sessionPool.Execute(builder.String())
	if err != nil {
		return err
	}
	builder.Reset()
	builder.WriteString("insert vertex user() values ")
	builder.WriteString(strconv.FormatInt(authorID, 10))
	builder.WriteString(":()")
	_, err = sessionPool.Execute(builder.String())
	if err != nil {
		return err
	}

	// 添加关注边
	builder.Reset()
	builder.WriteString("insert edge follow() values ")
	builder.WriteString(strconv.FormatInt(userID, 10))
	builder.WriteString("->")
	builder.WriteString(strconv.FormatInt(authorID, 10))
	builder.WriteString(":()")
	_, err = sessionPool.Execute(builder.String())

	return err
}

func UnFollow(ctx context.Context, userID, authorID int64) error {
	var builder strings.Builder
	builder.WriteString("delete edge follow ")
	builder.WriteString(strconv.FormatInt(userID, 10))
	builder.WriteString("->")
	builder.WriteString(strconv.FormatInt(authorID, 10))
	_, err := sessionPool.Execute(builder.String())

	return err
}

func FollowList(ctx context.Context, userID int64) (followList []int64, err error) {
	// 使用singleflight防止缓存击穿并减少redis压力
	key := GetRedisKey(KeyUserFollowPF, strconv.FormatInt(userID, 10))
	_, err, _ = g.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(delayTime)
			g.Forget(key)
		}()

		// 先查询redis缓存
		userIDs, err := RDB.SMembers(ctx, key).Result()
		if err == redis.Nil {
			// 缓存未命中, 查询数据库
			var builder strings.Builder
			builder.WriteString("match (v:user)-[:follow]->(v2:user) where id(v) == ")
			builder.WriteString(strconv.FormatInt(userID, 10))
			builder.WriteString(" return id(v2) as followList")
			resp, err := sessionPool.Execute(builder.String())
			if err != nil {
				return nil, err
			}

			res, err := resp.GetValuesByColName("followList")
			if err != nil {
				return nil, err
			}

			followList = make([]int64, len(res))
			for i, id := range res {
				followList[i], _ = id.AsInt()
			}

			// 写入redis缓存
			if len(followList) > 0 {
				RDB.SAdd(ctx, key, followList)
				RDB.Expire(ctx, key, ExpireTime+GetRandomTime())
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
	key := GetRedisKey(KeyUserFollowerPF, strconv.FormatInt(userID, 10))
	_, err, _ = g.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(delayTime)
			g.Forget(key)
		}()

		// 先查询redis缓存
		userIDs, err := RDB.SMembers(ctx, key).Result()
		if err == redis.Nil {
			// 缓存未命中，查询mysql
			var builder strings.Builder
			builder.WriteString("match (v:user)<-[:follow]-(v2:user) where id(v) == ")
			builder.WriteString(strconv.FormatInt(userID, 10))
			builder.WriteString(" return id(v2) as followerList limit 50")
			resp, err := sessionPool.Execute(builder.String())
			if err != nil {
				return nil, err
			}

			res, err := resp.GetValuesByColName("followerList")
			if err != nil {
				return nil, err
			}

			followerList = make([]int64, len(res))
			for i, id := range res {
				followerList[i], _ = id.AsInt()
			}

			// 写入redis缓存
			if len(followerList) > 0 {
				RDB.SAdd(ctx, key, followerList[:min(len(followerList), 50)])
				RDB.Expire(ctx, key, ExpireTime+GetRandomTime())
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

func FriendList(ctx context.Context, userID int64) (friendList []int64, err error) {
	// 使用singleflight解决缓存击穿并减少redis压力
	key := GetRedisKey(KeyUserFriendPF, strconv.FormatInt(userID, 10))
	_, err, _ = g.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(delayTime)
			g.Forget(key)
		}()

		// 先查询redis缓存
		userIDs, err := RDB.SMembers(ctx, key).Result()
		if err == redis.Nil {
			// 缓存未命中，查询mysql
			var builder strings.Builder
			builder.WriteString("match (v:user)-[:follow]->(v2:user)-[:follow]->(v:user) where id(v) == ")
			builder.WriteString(strconv.FormatInt(userID, 10))
			builder.WriteString(" return id(v2) as friendList")
			resp, err := sessionPool.Execute(builder.String())
			if err != nil {
				return nil, err
			}

			res, err := resp.GetValuesByColName("friendList")
			if err != nil {
				return nil, err
			}

			friendList = make([]int64, len(res))
			for i, id := range res {
				friendList[i], _ = id.AsInt()
			}

			return nil, nil
		} else if err != nil {
			return nil, err
		} else {
			// 缓存命中，转换为int64
			friendList = make([]int64, len(userIDs))
			for i, id := range userIDs {
				userID, err := strconv.ParseInt(id, 10, 64)
				if err != nil {
					return nil, err
				}
				friendList[i] = userID
			}

			return nil, nil
		}
	})

	return
}

func GetUserFollowCount(ctx context.Context, userID int64) (cnt int64, err error) {
	// 使用singleflight解决缓存击穿并减少redis压力
	key := GetRedisKey(KeyUserFollowCountPF, strconv.FormatInt(userID, 10))
	_, err, _ = g.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(delayTime)
			g.Forget(key)
		}()

		// 先查询redis缓存
		cnt, err = RDB.Get(ctx, key).Int64()
		if err == redis.Nil {
			// 缓存未命中，查询mysql
			var builder strings.Builder
			builder.WriteString("match (v:user)-[:follow]->(v2:user) where id(v) == ")
			builder.WriteString(strconv.FormatInt(userID, 10))
			builder.WriteString(" return count(*) as followCount")
			resp, err := sessionPool.Execute(builder.String())
			if err != nil {
				return nil, err
			}

			res, err := resp.GetValuesByColName("followCount")
			if err != nil {
				return nil, err
			}

			cnt, _ = res[0].AsInt()

			// 写入redis缓存
			err = RDB.Set(ctx, key, cnt, 0).Err()
			return nil, err
		}
		return nil, err
	})

	return
}

func GetUserFollowerCount(ctx context.Context, userID int64) (cnt int64, err error) {
	// 使用singleflight解决缓存击穿并减少redis压力
	key := GetRedisKey(KeyUserFollowerCountPF, strconv.FormatInt(userID, 10))
	_, err, _ = g.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(delayTime)
			g.Forget(key)
		}()

		// 先查询redis缓存
		cnt, err = RDB.Get(ctx, key).Int64()
		if err == redis.Nil {
			// 缓存未命中，查询mysql
			var builder strings.Builder
			builder.WriteString("match (v:user)<-[:follow]-(v2:user) where id(v)==")
			builder.WriteString(strconv.FormatInt(userID, 10))
			builder.WriteString(" return count(*) as followerCount")
			resp, err := sessionPool.Execute(builder.String())
			if err != nil {
				return nil, err
			}

			res, err := resp.GetValuesByColName("followerCount")
			if err != nil {
				return nil, err
			}

			cnt, _ = res[0].AsInt()

			// 写入redis缓存
			err = RDB.Set(ctx, key, cnt, 0).Err()
			return nil, err
		}

		return nil, err
	})

	return
}
