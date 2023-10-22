package service

import (
	"douyin/dao/mysql"
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

func UserInfo(authorID, userID int64) (*response.UserInfoResponse, error) {
	// 查询用户信息
	user, err := mysql.GetUserByID(userID, authorID)
	if err != nil {
		hlog.Error("service.UserInfo: 查询用户信息失败")
		return nil, err
	}

	// 返回响应
	return &response.UserInfoResponse{
		Response: &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
		User:     response.ToUserResponse(user),
	}, nil
}

func Register(username, password string) (*response.RegisterResponse, error) {
	// 查询用户是否已存在
	user := mysql.GetUserByName(username)
	if user != nil {
		hlog.Error("service.Register: 用户已存在")
		return nil, mysql.ErrUserExist
	}

	// 生成用户id
	userID := snowflake.GenerateID()

	// 加密密码
	bPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		hlog.Error("service.Register: 加密密码失败")
		return nil, ErrBcrypt
	}
	password = string(bPassword)

	// 保存用户信息
	mysql.CreateUser(username, password, userID)
	if err != nil {
		hlog.Error("service.Register: 保存用户信息失败")
		return nil, err
	}

	// 生成用户token
	token, err := jwt.GenerateToken(userID)
	if err != nil {
		hlog.Error("service.Register: 生成用户token失败")
		return nil, err
	}

	// 返回响应
	return &response.RegisterResponse{
		Response: &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
		UserID:   userID,
		Token:    token,
	}, nil
}

func Login(username, password string) (*response.LoginResponse, error) {
	// 查询用户是否存在
	user := mysql.GetUserByName(username)
	if user == nil {
		hlog.Error("service.Login: 用户不存在")
		return nil, mysql.ErrUserNotExist
	}

	// 校验密码
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		hlog.Error("service.Login: 校验密码失败")
		return nil, err
	}

	// 生成用户token
	token, err := jwt.GenerateToken(user.ID)
	if err != nil {
		hlog.Error("service.Login: 生成用户token失败")
		return nil, err
	}

	// 返回响应
	return &response.LoginResponse{
		Response: &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
		UserID:   user.ID,
		Token:    token,
	}, nil
}
