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

func PublishAction(ctx *app.RequestContext, userID int64, data *multipart.FileHeader, title string) (*response.Response, error) {
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

	// 保存视频信息到数据库
	err = mysql.SaveVideo(userID, data.Filename, coverName, title)
	if err != nil {
		hlog.Error("service.PublishAction: 保存视频信息到数据库失败, err: ", err)
		return nil, err
	}
	return &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()}, nil
}
