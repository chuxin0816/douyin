package oss

import (
	"bytes"
	"context"
	"douyin/config"
	"douyin/pkg/tracing"
	"io"
	"os"
	"path"
	"sync"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/cloudwego/kitex/pkg/klog"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"go.opentelemetry.io/otel/codes"
)

var bucket *oss.Bucket

const (
	pathName  = "./public/"
	videoPath = "video/"
	imagePath = "image/"
)

func Init() {
	tracing.Init(context.Background(), config.Conf.OpenTelemetryConfig.OssName)
	defer tracing.Close()

	client, err := oss.New(config.Conf.OssConfig.Endpoint, config.Conf.OssConfig.AccessKeyId, config.Conf.OssConfig.AccessKeySecret)
	if err != nil {
		panic(err)
	}
	bucket, err = client.Bucket(config.Conf.OssConfig.BucketName)
	if err != nil {
		panic(err)
	}
}

// UploadFile 上传文件到oss
func UploadFile(data []byte, uuidName string) error {
	_, span := tracing.Tracer.Start(context.Background(), "UploadFile")
	defer span.End()

	videoName := uuidName + ".mp4"
	coverName := uuidName + ".jpeg"

	// 使用协程并发执行文件上传操作
	var uploadErr error
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		// 上传视频
		if err := bucket.PutObject(path.Join(videoPath, videoName), bytes.NewReader(data)); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "上传视频失败")
			klog.Error("上传视频失败", err)
			uploadErr = err
		}
	}()

	go func() {
		defer wg.Done()
		// 获取视频封面
		imageData, err := getCoverImage(videoName)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "获取封面失败")
			klog.Error("获取封面失败, err: ", err)
			uploadErr = err
			return
		}
		if err := bucket.PutObject(path.Join(imagePath, coverName), imageData); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "上传封面失败")
			klog.Error("上传封面失败", err)
			uploadErr = err
		}
	}()

	wg.Wait()

	// 删除本地视频
	go func() {
		if err := os.Remove(videoName); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "删除视频失败")
			klog.Error("删除视频失败, err: ", err)
		}
	}()

	return uploadErr
}

// getCoverImage 获取视频第15帧作为封面
func getCoverImage(videoName string) (io.Reader, error) {
	_, span := tracing.Tracer.Start(context.Background(), "GetCoverImage")
	defer span.End()

	buf := bytes.NewBuffer(nil)
	err := ffmpeg.Input(videoName).Filter("select", ffmpeg.Args{"gte(n,15)"}).
		Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
		WithOutput(buf, nil).
		Run()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取封面失败")
		klog.Error("获取封面失败, err: ", err)
		return nil, err
	}
	return bytes.NewReader(buf.Bytes()), nil
}
