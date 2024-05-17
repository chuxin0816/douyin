package dal

import (
	"context"
	"strconv"
	"sync"
	"time"

	"douyin/src/dal/model"
	"douyin/src/dal/query"
	"douyin/src/kitex_gen/user"
	"douyin/src/pkg/snowflake"

	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
	"gorm.io/gorm"
)

// GetUserByID 用户通过作者id查询作者信息
func GetUserByID(ctx context.Context, authorID int64) (user *model.User, err error) {
	// 先判断布隆过滤器中是否存在
	if !bloomFilter.Test([]byte(strconv.FormatInt(authorID, 10))) {
		return nil, ErrUserNotExist
	}

	// 使用singleflight解决缓存击穿并减少redis压力
	key := GetRedisKey(KeyUserInfoPF, strconv.FormatInt(authorID, 10))
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
			b, err := msgpack.Marshal(user)
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
			err = msgpack.Unmarshal([]byte(userInfo), user)
			return nil, err
		}
	})

	return
}

// GetUserByIDs 根据用户id列表查询用户信息
func GetUserByIDs(ctx context.Context, authorIDs []int64) ([]*model.User, error) {
	users := make([]*model.User, len(authorIDs))

	for i, authorID := range authorIDs {
		user, err := GetUserByID(ctx, authorID)
		if err != nil {
			return nil, err
		}
		users[i] = user
	}

	return users, nil
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

func ToUserResponse(ctx context.Context, followerID *int64, mUser *model.User) *user.User {
	userResponse := &user.User{
		Id:              mUser.ID,
		Name:            mUser.Name,
		Avatar:          &mUser.Avatar,
		BackgroundImage: &mUser.BackgroundImage,
		IsFollow:        false,
		Signature:       &mUser.Signature,
	}

	var wg sync.WaitGroup
	var wgErr error
	wg.Add(5)
	go func() {
		defer wg.Done()
		cnt, err := GetUserFavoriteCount(ctx, mUser.ID)
		if err != nil {
			wgErr = err
			return
		}
		userResponse.FavoriteCount = &cnt
	}()
	go func() {
		defer wg.Done()
		cnt, err := GetUserTotalFavorited(ctx, mUser.ID)
		if err != nil {
			wgErr = err
			return
		}
		userResponse.TotalFavorited = &cnt
	}()
	go func() {
		defer wg.Done()
		cnt, err := GetUserFollowCount(ctx, mUser.ID)
		if err != nil {
			wgErr = err
			return
		}
		userResponse.FollowCount = &cnt
	}()
	go func() {
		defer wg.Done()
		cnt, err := GetUserFollowerCount(ctx, mUser.ID)
		if err != nil {
			wgErr = err
			return
		}
		userResponse.FollowerCount = &cnt
	}()
	go func() {
		defer wg.Done()
		cnt, err := GetUserWorkCount(ctx, mUser.ID)
		if err != nil {
			wgErr = err
			return
		}
		userResponse.WorkCount = &cnt
	}()
	wg.Wait()
	if wgErr != nil {
		return userResponse
	}

	// 判断是否关注
	if followerID == nil || *followerID == 0 {
		return userResponse
	}
	exist, err := CheckRelationExist(ctx, *followerID, mUser.ID)
	if err != nil {
		return userResponse
	}
	userResponse.IsFollow = exist

	return userResponse
}
