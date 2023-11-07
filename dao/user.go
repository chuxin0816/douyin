package dao

import (
	"douyin/models"
	"douyin/response"
	"errors"
	"strconv"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

var (
	ErrUserExist    = errors.New("用户已存在")
	ErrUserNotExist = errors.New("用户不存在")
	ErrPassword     = errors.New("密码错误")
)

// GetUserByID 用户通过作者id查询作者信息
func GetUserByID(authorID int64) (*models.User, error) {
	// 先判断布隆过滤器中是否存在
	if !bloomFilter.Test([]byte(strconv.FormatInt(authorID, 10))) {
		return nil, ErrUserNotExist
	}

	user := &models.User{}
	err := db.Where("id = ?", authorID).Find(user).Error
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, ErrUserNotExist
	}

	return user, nil
}

// GetUserByIDs 根据用户id列表查询用户信息
func GetUserByIDs(authorIDs []int64) ([]*models.User, error) {
	// 先判断布隆过滤器中是否存在
	for _, id := range authorIDs {
		if !bloomFilter.Test([]byte(strconv.FormatInt(id, 10))) {
			hlog.Error("mysql.GetUserByIDs: 用户不存在,id: ", id)
			return nil, ErrUserNotExist
		}
	}
	// 查询数据库
	var users []*models.User
	err := db.Where("id IN (?)", authorIDs).Find(&users).Error
	if err != nil {
		hlog.Error("mysql.GetUserByIDs: 查询数据库失败")
		return nil, err
	}

	// 解决重复字段缺少问题
	userMap := make(map[int64]*models.User)
	for _, dUser := range users {
		userMap[dUser.ID] = dUser
	}

	// 将users按照ids的顺序排列
	users = make([]*models.User, 0, len(authorIDs))
	for _, id := range authorIDs {
		user := userMap[id]
		users = append(users, user)
	}

	return users, nil
}

// GetUserByName 根据用户名查询用户信息, 如果用户不存在则返回nil
func GetUserByName(username string) *models.User {
	// 先判断布隆过滤器中是否存在
	if !bloomFilter.Test([]byte(username)) {
		return nil
	}

	user := &models.User{}
	db.Where("name = ?", username).Find(user)
	if user.ID == 0 {
		return nil
	}
	return user
}

func CreateUser(username, password string, userID int64) error {
	// 写入布隆过滤器
	bloomFilter.Add([]byte(strconv.FormatInt(userID, 10)))
	bloomFilter.Add([]byte(username))

	user := &models.User{
		ID:       userID,
		Name:     username,
		Password: password,
	}
	err := db.Create(user).Error
	if err != nil {
		hlog.Error("mysql.CreateUser: 保存用户信息失败")
		return err
	}
	return nil
}

func ToUserResponse(userID int64, user *models.User) *response.UserResponse {
	userResponse := &response.UserResponse{
		ID:              user.ID,
		Name:            user.Name,
		Avatar:          user.Avatar,
		BackgroundImage: user.BackgroundImage,
		FavoriteCount:   user.FavoriteCount,
		FollowCount:     user.FollowCount,
		FollowerCount:   user.FollowerCount,
		WorkCount:       user.WorkCount,
		IsFollow:        false,
		Signature:       user.Signature,
		TotalFavorited:  user.TotalFavorited,
	}

	// 判断是否关注
	// 从缓存中查询是否关注
	relation := &models.Relation{}
	db.Where("user_id = ? AND follower_id = ?", user.ID, userID).Find(relation)
	if relation.ID != 0 {
		userResponse.IsFollow = true
	}

	return userResponse
}
