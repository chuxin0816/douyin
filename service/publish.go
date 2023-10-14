package service

import (
	"douyin/pkg/ffmpeg"
	"douyin/pkg/oss"
	"douyin/response"
	"mime/multipart"
	"os"
	"path"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/google/uuid"
)

const (
	pathName = "./public/"
)

func PublishAction(ctx *app.RequestContext, userID int64, data *multipart.FileHeader, title string) (*response.Response, error) {
	// 保存视频到本地
	if err := os.MkdirAll(pathName, 0o777); err != nil {
		hlog.Error("service.PublishAction: 创建目录失败, err: ", err)
		return nil, err
	}
	uuidName := uuid.New().String()
	data.Filename = uuidName + ".mp4"
	videoPath := path.Join(pathName, data.Filename)
	if err := ctx.SaveUploadedFile(data, videoPath); err != nil {
		hlog.Error("service.PublishAction: 保存视频失败, err: ", err)
		return nil, err
	}

	// 获取视频封面
	coverName := uuidName + ".jpg"
	coverPath := path.Join(pathName, coverName)
	if err := ffmpeg.GetCoverImage(videoPath, coverPath); err != nil {
		hlog.Error("service.PublishAction: 获取封面失败, err: ", err)
		return nil, err
	}

	// 上传视频和封面到oss
	err := oss.UploadFile(data.Filename, coverName)
	if err != nil {
		hlog.Error("service.PublishAction: 上传文件失败, err: ", err)
		return nil, err
	}

	// 删除本地视频和封面
	go func() {
		if err := os.Remove(videoPath); err != nil {
			hlog.Error("service.PublishAction: 删除视频失败, err: ", err)
		}
		if err := os.Remove(coverPath); err != nil {
			hlog.Error("service.PublishAction: 删除封面失败, err: ", err)
		}
	}()
	// 保存视频信息到数据库
	return nil, nil
}
