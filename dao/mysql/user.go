package mysql

import (
	"douyin/models"
	"errors"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

var (
	ErrUserExist    = errors.New("用户已存在")
	ErrUserNotExist = errors.New("用户不存在")
	ErrPassword     = errors.New("密码错误")
)

// GetUserByID 用户通过作者id查询作者信息
func GetUserByID(userID, authorID int64) (*models.User, error) {
	user := &models.User{}
	err := db.Where("id = ?", authorID).Find(user).Error
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, ErrUserNotExist
	}

	// TODO: 通过用户id查询数据库判断是否关注

	return user, nil
}

// GetUserByIDs 根据用户id列表查询用户信息
func GetUserByIDs(userID int64, authorIDs []int64) ([]*models.User, error) {
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
		user, ok := userMap[id]
		if !ok {
			hlog.Error("mysql.GetUserByIDs: 用户不存在,id: ", id)
			return nil, ErrUserNotExist
		}

		// TODO: 通过用户id查询数据库判断是否关注

		users = append(users, user)
	}
	return users, nil
}

// GetUserByName 根据用户名查询用户信息, 如果用户不存在则返回nil
func GetUserByName(username string) *models.User {
	user := &models.User{}
	db.Where("name = ?", username).Find(user)
	if user.ID == 0 {
		return nil
	}
	return user
}

func CreateUser(req *models.UserRequest, userID int64) error {
	user := &models.User{
		ID:       userID,
		Name:     req.Username,
		Password: req.Password,
	}
	err := db.Create(user).Error
	if err != nil {
		hlog.Error("mysql.CreateUser: 保存用户信息失败")
		return err
	}
	return nil
}
