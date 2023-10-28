package oss

import (
	"bytes"
	"douyin/config"
	"io"
	"mime/multipart"
	"os"
	"path"
	"sync"

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

// UploadFile 上传文件到oss
func UploadFile(file *multipart.FileHeader, uuidName string) error {
	videoName := uuidName + ".mp4"
	coverName := uuidName + ".jpeg"
	
	// 使用协程并发执行文件上传操作
	var uploadErr error
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		// 打开视频获取视频流
		videoStream, err := file.Open()
		if err != nil {
			hlog.Error("oss.UploadFile: 打开视频失败, err: ", err)
			uploadErr = err
			return
		}

		// 上传视频
		if err := bucket.PutObject(path.Join(videoPath, videoName), videoStream); err != nil {
			hlog.Error("oss.UploadFile: 上传视频失败", err)
			uploadErr = err
		}
	}()

	go func() {
		defer wg.Done()
		// 获取视频封面
		imageData, err := GetCoverImage(videoName)
		if err != nil {
			hlog.Error("oss.UploadFile: 获取封面失败, err: ", err)
			uploadErr = err
			return
		}
		if err := bucket.PutObject(path.Join(imagePath, coverName), imageData); err != nil {
			hlog.Error("oss.UploadFile: 上传封面失败", err)
			uploadErr = err
		}
	}()

	wg.Wait()

	// 删除本地视频
	go func() {
		if err := os.Remove(videoName); err != nil {
			hlog.Error("service.PublishAction: 删除视频失败, err: ", err)
		}
	}()

	return uploadErr
}

// GetCoverImage 获取视频第15帧作为封面
func GetCoverImage(videoName string) (io.Reader, error) {
	buf := bytes.NewBuffer(nil)
	err := ffmpeg.Input(videoName).Filter("select", ffmpeg.Args{"gte(n,15)"}).
		Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
		WithOutput(buf, nil).
		Run()
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf.Bytes()), nil
}
