package main

import (
	"context"
	"sync"

	"douyin/src/client"
	"douyin/src/common/jwt"
	"douyin/src/dal"
	"douyin/src/kitex_gen/user"

	"github.com/cloudwego/kitex/pkg/klog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/crypto/bcrypt"
)

// UserServiceImpl implements the last service interface defined in the IDL.
type UserServiceImpl struct{}

// Register implements the UserServiceImpl interface.
func (s *UserServiceImpl) Register(ctx context.Context, req *user.UserRegisterRequest) (resp *user.UserRegisterResponse, err error) {
	ctx, span := otel.Tracer("user").Start(ctx, "Register")
	defer span.End()

	// 查询用户是否已存在
	mUser := dal.GetUserLoginByName(ctx, req.Username)
	if mUser != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "用户已存在")
		klog.Error("用户已存在")
		return nil, dal.ErrUserExist
	}

	// 加密密码
	bPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "加密密码失败")
		klog.Error("加密密码失败")
		return nil, err
	}
	req.Password = string(bPassword)

	// 保存用户信息
	userID, err := dal.CreateUser(ctx, req.Username, req.Password)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "保存用户信息失败")
		klog.Error("保存用户信息失败")
		return nil, err
	}

	// 生成用户token
	token, err := jwt.GenerateToken(userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "生成用户token失败")
		klog.Error("生成用户token失败")
		return nil, err
	}

	// 返回响应
	resp = &user.UserRegisterResponse{UserId: userID, Token: token}
	return
}

// Login implements the UserServiceImpl interface.
func (s *UserServiceImpl) Login(ctx context.Context, req *user.UserLoginRequest) (resp *user.UserLoginResponse, err error) {
	ctx, span := otel.Tracer("user").Start(ctx, "Login")
	defer span.End()

	// 查询用户是否存在
	mUser := dal.GetUserLoginByName(ctx, req.Username)
	if mUser == nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "用户不存在")
		klog.Error("用户不存在")
		return nil, dal.ErrUserNotExist
	}

	// 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(mUser.Password), []byte(req.Password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			klog.Info("密码错误, username: ", req.Username)
			return nil, dal.ErrPassword
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "校验密码失败")
		klog.Error("校验密码失败")
		return nil, err
	}

	// 生成用户token
	token, err := jwt.GenerateToken(mUser.ID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "生成用户token失败")
		klog.Error("生成用户token失败")
		return nil, err
	}

	// 返回响应
	resp = &user.UserLoginResponse{UserId: mUser.ID, Token: token}

	return
}

// UserInfo implements the UserServiceImpl interface.
func (s *UserServiceImpl) UserInfo(ctx context.Context, req *user.UserInfoRequest) (resp *user.UserInfoResponse, err error) {
	ctx, span := otel.Tracer("user").Start(ctx, "UserInfo")
	defer span.End()

	// 查询用户信息
	mUser, err := dal.GetUserByID(ctx, req.AuthorId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "查询用户信息失败")
		klog.Error("查询用户信息失败")
		return nil, err
	}

	userResponse := &user.User{
		Id:              mUser.ID,
		Name:            mUser.Name,
		Avatar:          mUser.Avatar,
		BackgroundImage: mUser.BackgroundImage,
		IsFollow:        false,
		Signature:       mUser.Signature,
	}
	var wg sync.WaitGroup
	var wgErr error
	wg.Add(5)
	go func() {
		defer wg.Done()
		cnt, err := client.FavoriteClient.FavoriteCnt(ctx, mUser.ID)
		if err != nil {
			wgErr = err
			return
		}
		userResponse.FavoriteCount = cnt
	}()
	go func() {
		defer wg.Done()
		cnt, err := client.FavoriteClient.TotalFavoritedCnt(ctx, mUser.ID)
		if err != nil {
			wgErr = err
			return
		}
		userResponse.TotalFavorited = cnt
	}()
	go func() {
		defer wg.Done()
		cnt, err := client.RelationClient.FollowCnt(ctx, mUser.ID)
		if err != nil {
			wgErr = err
			return
		}
		userResponse.FollowCount = cnt
	}()
	go func() {
		defer wg.Done()
		cnt, err := client.RelationClient.FollowerCnt(ctx, mUser.ID)
		if err != nil {
			wgErr = err
			return
		}
		userResponse.FollowerCount = cnt
	}()
	go func() {
		defer wg.Done()
		cnt, err := client.VideoClient.WorkCount(ctx, mUser.ID)
		if err != nil {
			wgErr = err
			return
		}
		userResponse.WorkCount = cnt
	}()
	wg.Wait()
	if wgErr != nil {
		span.RecordError(wgErr)
		span.SetStatus(codes.Error, "查询用户信息失败")
		klog.Error("查询用户信息失败")
		return nil, wgErr
	}

	// 判断是否关注
	if req.UserId == nil {
		return
	}
	exist, err := client.RelationClient.RelationExist(ctx, *req.UserId, mUser.ID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "查询用户信息失败")
		klog.Error("查询用户信息失败")
		return nil, err
	}
	userResponse.IsFollow = exist

	// 返回响应
	resp = &user.UserInfoResponse{User: userResponse}

	return
}
