package service

import (
	"douyin/dao/mysql"
	"douyin/models"
	"douyin/pkg/jwt"
	"douyin/pkg/snowflake"
	"douyin/response"
	"errors"
	"strconv"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExist    = errors.New("用户已存在")
	ErrUserNotExist = errors.New("用户不存在")
	ErrPassword     = errors.New("密码错误")
	ErrBcrypt       = errors.New("加密密码失败")
)

func UserInfo(req *models.UserInfoRequest) (*response.UserInfoResponse, error) {
	// 查询用户信息
	user := mysql.GetUserByID(strconv.FormatInt(req.UserID, 10))
	if user == nil {
		hlog.Error("service.UserInfo: 用户不存在")
		return nil, ErrUserNotExist
	}

	// 返回响应
	return &response.UserInfoResponse{
		Response: &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
		User:     response.ToUserResponse(user),
	}, nil
}

func Register(req *models.UserRequest) (*response.RegisterResponse, error) {
	// 查询用户是否已存在
	user := mysql.GetUserByName(req.Username)
	if user != nil {
		hlog.Error("service.Register: 用户已存在")
		return nil, ErrUserExist
	}

	// 生成用户id
	userID := snowflake.GenerateID()

	// 加密密码
	password, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		hlog.Error("service.Register: 加密密码失败")
		return nil, ErrBcrypt
	}
	req.Password = string(password)

	// 保存用户信息
	mysql.CreateUser(req)
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

func Login(req *models.UserRequest) (*response.LoginResponse, error) {
	// 查询用户是否存在
	user := mysql.GetUserByName(req.Username)
	if user == nil {
		hlog.Error("service.Login: 用户不存在")
		return nil, ErrUserNotExist
	}

	// 校验密码
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, ErrPassword
		}
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
