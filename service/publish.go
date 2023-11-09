package service

import (
	"douyin/dao"
	"douyin/pkg/oss"
	"douyin/response"
	"mime/multipart"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/google/uuid"
)

func PublishAction(ctx *app.RequestContext, userID int64, data *multipart.FileHeader, title string) (*response.PublishActionResponse, error) {
	// 生成uuid作为文件名
	uuidName := uuid.New().String()
	data.Filename = uuidName + ".mp4"
	coverName := uuidName + ".jpeg"

	// 保存视频到本地
	if err := ctx.SaveUploadedFile(data, data.Filename); err != nil {
		hlog.Error("service.PublishAction: 保存视频失败, err: ", err)
		return nil, err
	}

	// 上传视频和封面到oss
	go func() {
		if err := oss.UploadFile(data, uuidName); err != nil {
			hlog.Error("service.PublishAction: 上传文件失败, err: ", err)
		}
	}()

	// 操作数据库
	if err := dao.SaveVideo(userID, data.Filename, coverName, title); err != nil {
		hlog.Error("service.PublishAction: 操作数据库失败, err: ", err)
		return nil, err
	}

	return &response.PublishActionResponse{
		Response: &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
	}, nil
}

func PublishList(userID, authorID int64) (*response.PublishListResponse, error) {
	// 查询视频列表
	resp, err := dao.GetPublishList(userID, authorID)
	if err != nil {
		hlog.Error("service.PublishList: 查询视频列表失败, err: ", err)
		return nil, err
	}

	// 返回响应
	return &response.PublishListResponse{
		Response:  &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
		VideoList: resp,
	}, nil
}
