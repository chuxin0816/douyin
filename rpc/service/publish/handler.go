package main

import (
	"context"
	"douyin/dal"
	"douyin/pkg/oss"
	publish "douyin/rpc/kitex_gen/publish"
	"os"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/google/uuid"
)

// PublishServiceImpl implements the last service interface defined in the IDL.
type PublishServiceImpl struct{}

// PublishAction implements the PublishServiceImpl interface.
func (s *PublishServiceImpl) PublishAction(ctx context.Context, req *publish.PublishActionRequest) (resp *publish.PublishActionResponse, err error) {
	// 生成uuid作为文件名
	uuidName := uuid.New().String()
	videoName := uuidName + ".mp4"
	coverName := uuidName + ".jpeg"

	// 保存视频到本地
	if err := os.WriteFile(videoName, req.Data, 0644); err != nil {
		klog.Error("保存视频到本地失败, err: ", err)
		return nil, err
	}

	// 保存视频到本地
	go func() {
		if err := oss.UploadFile(req.Data, uuidName); err != nil {
			klog.Error("保存视频到oss失败, err: ", err)
		}
	}()

	// 操作数据库
	if err := dal.SaveVideo(req.UserId, videoName, coverName, req.Title); err != nil {
		klog.Error("操作数据库失败, err: ", err)
		return nil, err
	}

	// 返回响应
	resp = &publish.PublishActionResponse{}

	return
}

// PublishList implements the PublishServiceImpl interface.
func (s *PublishServiceImpl) PublishList(ctx context.Context, req *publish.PublishListRequest) (resp *publish.PublishListResponse, err error) {
	// 查询视频列表
	videoList, err := dal.GetPublishList(req.UserId, req.ToUserId)
	if err != nil {
		klog.Error("查询视频列表失败, err: ", err)
		return nil, err
	}

	// 返回响应
	resp = &publish.PublishListResponse{VideoList: videoList}

	return
}
