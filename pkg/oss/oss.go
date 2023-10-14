package oss

import (
	"bytes"
	"douyin/config"
	"io"
	"os"
	"path"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	ffmpeg "github.com/u2takey/ffmpeg-go"
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

func UploadFile(uuidName string) error {
	// 获取视频封面
	videoName := uuidName + ".mp4"
	coverName := uuidName + ".jpeg"
	imageData, err := GetCoverImage(videoName)
	if err != nil {
		hlog.Error("service.PublishAction: 获取封面失败, err: ", err)
		return err
	}

	// 上传视频
	err = bucket.PutObjectFromFile(path.Join(videoPath, videoName), videoName)
	if err != nil {
		hlog.Error("oss.UploadFile: 上传视频失败", err)
		return err
	}

	// 删除本地视频和封面
	go func() {
		if err := os.Remove(videoName); err != nil {
			hlog.Error("service.PublishAction: 删除视频失败, err: ", err)
		}
	}()

	// 上传封面
	err = bucket.PutObject(path.Join(imagePath, coverName), imageData)
	if err != nil {
		hlog.Error("oss.UploadFile: 上传封面失败", err)
		return err
	}
	return nil
}

func GetCoverImage(videoName string) (io.Reader, error) {
	buf := bytes.NewBuffer(nil)
	err := ffmpeg.Input(videoName).Filter("select", ffmpeg.Args{"gte(n,0)"}).
		Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
		WithOutput(buf, nil).
		Run()
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf.Bytes()), nil
}
