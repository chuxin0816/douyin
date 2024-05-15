package dal

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"douyin/src/dal/model"
	"douyin/src/dal/query"
	"douyin/src/kitex_gen/user"
	"douyin/src/pkg/snowflake"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// GetUserByID 用户通过作者id查询作者信息
func GetUserByID(ctx context.Context, authorID int64) (user *model.User, err error) {
	// 先判断布隆过滤器中是否存在
	if !bloomFilter.Test([]byte(strconv.FormatInt(authorID, 10))) {
		return nil, ErrUserNotExist
	}

	// 使用singleflight解决缓存击穿并减少redis压力
	key := GetRedisKey(KeyUserInfoPF + strconv.FormatInt(authorID, 10))
	_, err, _ = g.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(delayTime)
			g.Forget(key)
		}()

		// 先查询redis缓存
		userInfo, err := RDB.Get(ctx, key).Result()
		if err == redis.Nil {
			// 缓存未命中，查询mysql
			user, err = qUser.WithContext(ctx).Where(qUser.ID.Eq(authorID)).First()
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					return nil, ErrUserNotExist
				}
				return nil, err
			}

			// 写入redis缓存
			b, err := json.Marshal(user)
			if err != nil {
				return nil, err
			}
			userInfo = string(b)
			err = RDB.Set(ctx, key, userInfo, ExpireTime+GetRandomTime()).Err()
			return nil, err
		} else if err != nil {
			return nil, err
		} else {
			// 缓存命中
			err = json.Unmarshal([]byte(userInfo), user)
			return nil, err
		}
	})

	return
}

// GetUserByIDs 根据用户id列表查询用户信息
func GetUserByIDs(ctx context.Context, authorIDs []int64) ([]*model.User, error) {
	users := make([]*model.User, len(authorIDs))

	var wg sync.WaitGroup
	var firstErr error
	wg.Add(len(authorIDs))
	for i, authorID := range authorIDs {
		go func(i int, authorID int64) {
			defer wg.Done()
			user, err := GetUserByID(ctx, authorID)
			if err != nil && firstErr == nil {
				firstErr = err
				return
			}
			users[i] = user
		}(i, authorID)
	}
	wg.Wait()

	return users, firstErr
}

// GetUserLoginByName 根据用户名查询用户密码, 如果用户不存在则返回nil
func GetUserLoginByName(ctx context.Context, username string) *model.UserLogin {
	// 先判断布隆过滤器中是否存在
	if !bloomFilter.Test([]byte(username)) {
		return nil
	}

	// user, err := qUser.WithContext(ctx).Where(qUser.Name.Eq(username)).Select(qUser.ID, qUser.Password).First()
	user, err := qUserLogin.WithContext(ctx).Select(qUserLogin.ID, qUserLogin.Password).Where(qUserLogin.Username.Eq(username)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return nil
	}

	return user
}

func CreateUser(ctx context.Context, username, password string) (userID int64, err error) {
	userID = snowflake.GenerateID()

	// 写入布隆过滤器
	bloomFilter.Add([]byte(strconv.FormatInt(userID, 10)))
	bloomFilter.Add([]byte(username))

	user := &model.User{
		ID:   userID,
		Name: username,
	}
	userLogin := &model.UserLogin{
		ID:       userID,
		Username: username,
		Password: password,
	}

	err = q.Transaction(func(tx *query.Query) error {
		if err := tx.User.WithContext(ctx).Create(user); err != nil {
			return err
		}
		if err := tx.UserLogin.WithContext(ctx).Create(userLogin); err != nil {
			return err
		}
		return nil
	})

	return
}

func GetUserFavoriteCount(ctx context.Context, userID int64) (int64, error) {
	var cnt int64
	err := qUser.WithContext(ctx).Where(qUser.ID.Eq(userID)).Select(qUser.FavoriteCount).Scan(&cnt)
	return cnt, err
}

func GetUserTotalFavorited(ctx context.Context, userID int64) (int64, error) {
	var cnt int64
	err := qUser.WithContext(ctx).Where(qUser.ID.Eq(userID)).Select(qUser.TotalFavorited).Scan(&cnt)
	return cnt, err
}

func GetUserFollowCount(ctx context.Context, userID int64) (int64, error) {
	var cnt int64
	err := qUser.WithContext(ctx).Where(qUser.ID.Eq(userID)).Select(qUser.FollowCount).Scan(&cnt)
	return cnt, err
}

func GetUserFollowerCount(ctx context.Context, userID int64) (int64, error) {
	var cnt int64
	err := qUser.WithContext(ctx).Where(qUser.ID.Eq(userID)).Select(qUser.FollowerCount).Scan(&cnt)
	return cnt, err
}

func GetUserWorkCount(ctx context.Context, userID int64) (int64, error) {
	var cnt int64
	err := qUser.WithContext(ctx).Where(qUser.ID.Eq(userID)).Select(qUser.WorkCount).Scan(&cnt)
	return cnt, err
}

func ToUserResponse(ctx context.Context, followerID *int64, mUser *model.User) *user.User {
	userResponse := &user.User{
		Id:              mUser.ID,
		Name:            mUser.Name,
		Avatar:          &mUser.Avatar,
		BackgroundImage: &mUser.BackgroundImage,
		FavoriteCount:   &mUser.FavoriteCount,
		FollowCount:     &mUser.FollowCount,
		FollowerCount:   &mUser.FollowerCount,
		WorkCount:       &mUser.WorkCount,
		IsFollow:        false,
		Signature:       &mUser.Signature,
		TotalFavorited:  &mUser.TotalFavorited,
	}

	if followerID == nil || *followerID == 0 {
		return userResponse
	}

	// 判断是否关注
	exist, err := CheckRelationExist(ctx, *followerID, mUser.ID)
	if err != nil {
		return userResponse
	}
	userResponse.IsFollow = exist

	return userResponse
}

func UpdateUser(ctx context.Context, user *model.User) error {
	_, err := qUser.WithContext(ctx).Where(qUser.ID.Eq(user.ID)).Updates(map[string]any{
		"total_favorited": user.TotalFavorited,
		"favorite_count":  user.FavoriteCount,
		"follow_count":    user.FollowCount,
		"follower_count":  user.FollowerCount,
		"work_count":      user.WorkCount,
	})
	return err
}
