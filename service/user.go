package service

import (
	"douyin/dao/mysql"
	"douyin/models"
	"douyin/pkg/jwt"
	"douyin/pkg/snowflake"
	"douyin/response"
	"errors"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrBcrypt = errors.New("加密密码失败")
)

func Register(req *models.UserRequest) (*response.RegisterResponse, error) {
	// 查询用户是否已存在
	err := mysql.CheckUsernameExist(req.Username)
	if err != nil {
		if errors.Is(err, mysql.ErrUserExist) {
			return nil, err
		}
		hlog.Error("service.Register: 查询用户是否已存在失败")
		return nil, err
	}

	// 生成用户id
	userID := snowflake.GenerateID()

	// 生成用户token
	token, err := jwt.GenerateToken(userID)
	if err != nil {
		hlog.Error("service.Register: 生成用户token失败")
		return nil, err
	}

	// 加密密码
	password, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		hlog.Error("service.Register: 加密密码失败")
		return nil, ErrBcrypt
	}
	req.Password = string(password)

	// 保存用户信息
	err = mysql.CreateUser(req)
	if err != nil {
		hlog.Error("service.Register: 保存用户信息失败")
		return nil, err
	}

	// 返回响应
	return &response.RegisterResponse{
		Response: &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
		UserID:   userID,
		Token:    token,
	}, nil
}
