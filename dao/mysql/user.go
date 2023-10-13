package mysql

import (
	"douyin/models"
	"douyin/response"
	"strings"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func GetUserByID(id int64) *models.User {
	user := &models.User{}
	db.Where("id = ?", id).Find(user)
	if user.ID == 0 {
		return nil
	}
	return user
}

func GetUserByIDs(ids []string) (users []*response.UserResponse, err error) {
	// 查询数据库
	var dUsers []*models.User
	err = db.Where("id IN (?)", ids).Order("FIELD(id," + strings.Join(ids, ",") + ")").Find(&dUsers).Error
	if err != nil {
		hlog.Error("mysql.GetUserByIDs: 查询数据库失败")
		return nil, err
	}

	// 将models.User转换为response.UserResponse
	for _, dUser := range dUsers {
		users = append(users, response.ToUserResponse(dUser))
	}
	return
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
