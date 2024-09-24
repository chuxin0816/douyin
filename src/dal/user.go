package dal

import (
	"context"
	"strconv"
	"time"

	"douyin/src/common/snowflake"
	"douyin/src/dal/model"
	"douyin/src/dal/query"

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
	_, err, _ = G.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(DelayTime)
			G.Forget(key)
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

// GetUserLoginByName 根据用户名查询用户密码, 如果用户不存在则返回nil
func GetUserLoginByName(ctx context.Context, username string) *model.UserLogin {
	// 先判断布隆过滤器中是否存在
	if !bloomFilter.Test([]byte(username)) {
		return nil
	}

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
