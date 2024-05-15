package main

import (
	"context"
	"os"
	"sync"

	"douyin/src/dal"
	"douyin/src/dal/model"
	"douyin/src/kitex_gen/feed"
	publish "douyin/src/kitex_gen/publish"
	"douyin/src/pkg/oss"
	"douyin/src/pkg/tracing"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
)

// PublishServiceImpl implements the last service interface defined in the IDL.
type PublishServiceImpl struct{}

// PublishAction implements the PublishServiceImpl interface.
func (s *PublishServiceImpl) PublishAction(ctx context.Context, req *publish.PublishActionRequest) (resp *publish.PublishActionResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "PublishAction")
	defer span.End()

	// 生成uuid作为文件名
	uuidName := uuid.New().String()
	videoName := uuidName + ".mp4"
	coverName := uuidName + ".jpeg"

	// 保存视频到本地
	if err := os.WriteFile(videoName, req.Data, 0o644); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "保存视频到本地失败")
		klog.Error("保存视频到本地失败, err: ", err)
		return nil, err
	}

	// 上传视频到oss
	go func() {
		if err := oss.UploadFile(ctx, req.Data, uuidName); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "保存视频到oss失败")
			klog.Error("保存视频到oss失败, err: ", err)
		}
	}()

	// 操作数据库
	if err := dal.SaveVideo(ctx, req.UserId, videoName, coverName, req.Title); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "操作数据库失败")
		klog.Error("操作数据库失败, err: ", err)
		return nil, err
	}

	// 返回响应
	resp = &publish.PublishActionResponse{}

	return
}

// PublishList implements the PublishServiceImpl interface.
func (s *PublishServiceImpl) PublishList(ctx context.Context, req *publish.PublishListRequest) (resp *publish.PublishListResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "PublishList")
	defer span.End()

	// 查询视频列表
	mVideoList, err := dal.GetPublishList(ctx, req.ToUserId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "查询视频列表失败")
		klog.Error("查询视频列表失败, err: ", err)
		return nil, err
	}

	// 将model.Video转换为feed.Video
	videoList := make([]*feed.Video, len(mVideoList))
	var wg sync.WaitGroup
	wg.Add(len(mVideoList))
	for i, mVideo := range mVideoList {
		go func(i int, mVideo *model.Video) {
			defer wg.Done()
			videoList[i] = dal.ToVideoResponse(ctx, req.UserId, mVideo)
		}(i, mVideo)
	}
	wg.Wait()

	// 返回响应
	resp = &publish.PublishListResponse{VideoList: videoList}

	return
}
