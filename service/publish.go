package service

import (
	"douyin/dao/mysql"
	"douyin/pkg/oss"
	"douyin/response"
	"mime/multipart"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/google/uuid"
)

func PublishAction(ctx *app.RequestContext, userID int64, data *multipart.FileHeader, title string) (*response.PublishActionResponse, error) {
	// 保存视频到本地
	uuidName := uuid.New().String()
	data.Filename = uuidName + ".mp4"
	coverName := uuidName + ".jpeg"
	if err := ctx.SaveUploadedFile(data, data.Filename); err != nil {
		hlog.Error("service.PublishAction: 保存视频失败, err: ", err)
		return nil, err
	}

	// 上传视频和封面到oss
	err := oss.UploadFile(uuidName)
	if err != nil {
		hlog.Error("service.PublishAction: 上传文件失败, err: ", err)
		return nil, err
	}

	// 操作数据库
	err = mysql.SaveVideo(userID, data.Filename, coverName, title)
	if err != nil {
		hlog.Error("service.PublishAction: 操作数据库失败, err: ", err)
		return nil, err
	}
	return &response.PublishActionResponse{
		Response: &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
	}, nil
}

func PublishList(userID, authorID int64) (*response.PublishListResponse, error) {
	// 查询视频列表
	resp, err := mysql.GetPublishList(userID, authorID)
	if err != nil {
		hlog.Error("service.PublishList: 查询视频列表失败, err: ", err)
		return nil, err
	}

	// TODO: 通过用户id查询数据库判断是否点赞

	// 返回响应
	return &response.PublishListResponse{
		Response:  &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
		VideoList: resp,
	}, nil
}
