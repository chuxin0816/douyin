package oss

import (
	"douyin/config"
	"path"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

var bucket *oss.Bucket

const (
	pathName  = "./public/"
	videoPath = "video/"
	imagePath = "image/"
)

func Init(conf *config.OssConfig) error {
	client, err := oss.New(conf.Endpoint, conf.AccessKeyId, conf.AccessKeySecret)
	if err != nil {
		return err
	}
	bucket, err = client.Bucket(conf.BucketName)
	return err
}

func UploadFile(videoName, coverName string) error {
	// 上传视频
	err := bucket.PutObjectFromFile(path.Join(videoPath, videoName), path.Join(pathName, videoName))
	if err != nil {
		hlog.Error("oss.UploadFile: 上传视频失败", err)
		return err
	}

	// 上传封面
	err = bucket.PutObjectFromFile(path.Join(imagePath, coverName), path.Join(pathName, coverName))
	if err != nil {
		hlog.Error("oss.UploadFile: 上传封面失败", err)
		return err
	}
	return nil
}
