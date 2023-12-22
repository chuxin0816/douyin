package main

import (
	"context"
	"douyin/dal"
	"douyin/pkg/jwt"
	"douyin/pkg/snowflake"
	user "douyin/rpc/kitex_gen/user"
	"errors"

	"github.com/u2takey/go-utils/klog"
	"golang.org/x/crypto/bcrypt"
)

var ErrBcrypt = errors.New("加密密码失败")

// UserServiceImpl implements the last service interface defined in the IDL.
type UserServiceImpl struct{}

// Register implements the UserServiceImpl interface.
func (s *UserServiceImpl) Register(ctx context.Context, req *user.UserRegisterRequest) (resp *user.UserRegisterResponse, err error) {
	// 查询用户是否已存在
	mUser := dal.GetUserByName(req.Username)
	if mUser != nil {
		klog.Error("用户已存在")
		return nil, dal.ErrUserExist
	}

	// 生成用户id
	userID := snowflake.GenerateID()

	// 加密密码
	bPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		klog.Error("加密密码失败")
		return nil, ErrBcrypt
	}
	req.Password = string(bPassword)

	// 保存用户信息
	dal.CreateUser(req.Username, req.Password, userID)
	if err != nil {
		klog.Error("保存用户信息失败")
		return nil, err
	}

	// 生成用户token
	token, err := jwt.GenerateToken(userID)
	if err != nil {
		klog.Error("生成用户token失败")
		return nil, err
	}

	// 返回响应
	resp = &user.UserRegisterResponse{UserId: userID, Token: token}
	return
}

// Login implements the UserServiceImpl interface.
func (s *UserServiceImpl) Login(ctx context.Context, req *user.UserLoginRequest) (resp *user.UserLoginResponse, err error) {
	// 查询用户是否存在
	mUser := dal.GetUserByName(req.Username)
	if mUser == nil {
		klog.Error("用户不存在")
		return nil, dal.ErrUserNotExist
	}

	// 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(mUser.Password), []byte(req.Password));err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			klog.Info("密码错误, username: ", req.Username)
			return nil, dal.ErrPassword
		}
		klog.Error("校验密码失败")
		return nil, err
	}

	// 生成用户token
	token, err := jwt.GenerateToken(mUser.ID)
	if err != nil {
		klog.Error("生成用户token失败")
		return nil, err
	}

	// 返回响应
	resp = &user.UserLoginResponse{UserId: mUser.ID, Token: token}

	return
}

// UserInfo implements the UserServiceImpl interface.
func (s *UserServiceImpl) UserInfo(ctx context.Context, req *user.UserInfoRequest) (resp *user.UserInfoResponse, err error) {
	// 查询用户信息
	mUser, err := dal.GetUserByID(req.ToUserId)
	if err != nil {
		klog.Error("查询用户信息失败")
		return nil, err
	}
	
	// 返回响应
	resp = &user.UserInfoResponse{User: dal.ToUserResponse(*req.UserId, mUser)}
	
	return
}
