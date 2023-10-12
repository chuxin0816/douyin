package mysql

import (
	"douyin/models"
	"douyin/response"
	"errors"
	"strings"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

var (
	ErrUserExist    = errors.New("用户已存在")
	ErrUserNotExist = errors.New("用户不存在")
	ErrPassword     = errors.New("密码错误")
)

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
		users = append(users, &response.UserResponse{
			Avatar:          dUser.Avatar,
			BackgroundImage: dUser.BackgroundImage,
			FavoriteCount:   dUser.FavoriteCount,
			FollowCount:     dUser.FollowCount,
			FollowerCount:   dUser.FollowerCount,
			ID:              dUser.ID,
			IsFollow:        false, // 需要登录后通过用户id查询数据库判断
			Name:            dUser.Name,
			Signature:       dUser.Signature,
			TotalFavorited:  dUser.TotalFavorited,
			WorkCount:       dUser.WorkCount,
		})
	}
	return
}

func CheckUsernameExist(username string) error {
	user := &models.User{}
	db.Where("name = ?", username).First(user)
	if user.ID != 0 {
		return ErrUserExist
	}
	return nil
}

func CreateUser(req *models.UserRequest) error {
	user := &models.User{
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
