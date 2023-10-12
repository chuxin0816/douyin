package service

import (
	"douyin/models"
	"douyin/pkg/jwt"
	"douyin/pkg/snowflake"
	"douyin/response"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func Register(req *models.UserRequest) (*response.RegisterResponse, error) {
	// 查询用户是否已存在

	// 生成用户id
	userID, err := snowflake.GenerateID()
	if err != nil {
		hlog.Error("service.Register: 生成用户id失败", err)
		return nil, err
	}

	// 生成用户token
	token, err := jwt.GenerateToken(userID)
	if err != nil {
		hlog.Error("service.Register: 生成用户token失败", err)
		return nil, err
	}

	// 保存用户信息

	// 返回响应
	return &response.RegisterResponse{
		Response: &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
		UserID:   userID,
		Token:    token,
	}, nil
}
