package main

import (
	"context"
	"douyin/dal"
	"douyin/pkg/oss"
	"douyin/pkg/tracing"
	publish "douyin/rpc/kitex_gen/publish"
	"os"
	"strconv"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
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
	if err := os.WriteFile(videoName, req.Data, 0644); err != nil {
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

	// 更新缓存相关字段
	go func() {
		// 修改用户发布视频数
		key := dal.GetRedisKey(dal.KeyUserWorkCountPF + strconv.FormatInt(req.UserId, 10))
		// 检查key是否存在
		if exist, err := dal.RDB.Exists(ctx, key).Result(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "检查key是否存在失败")
			klog.Error("检查key是否存在失败, err: ", err)
			return
		} else if exist == 0 {
			// 缓存不存在，查询数据库写入缓存
			cnt, err := dal.GetUserWorkCount(ctx, req.UserId)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "查询用户作品数失败")
				klog.Error("查询用户作品数失败, err: ", err)
				return
			}
			if err := dal.RDB.Set(ctx, key, cnt, redis.KeepTTL).Err(); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "写入缓存失败")
				klog.Error("写入缓存失败, err: ", err)
				return
			}
		}
		if err := dal.RDB.Incr(ctx, key).Err(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "增加用户作品数失败")
			klog.Error("增加用户作品数失败, err: ", err)
			return
		}

		// 写入待同步队列
		dal.CacheUserID.Store(req.UserId, struct{}{})

	}()

	// 返回响应
	resp = &publish.PublishActionResponse{}

	return
}

// PublishList implements the PublishServiceImpl interface.
func (s *PublishServiceImpl) PublishList(ctx context.Context, req *publish.PublishListRequest) (resp *publish.PublishListResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "PublishList")
	defer span.End()

	// 查询视频列表
	videoList, err := dal.GetPublishList(ctx, req.UserId, req.ToUserId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "查询视频列表失败")
		klog.Error("查询视频列表失败, err: ", err)
		return nil, err
	}

	// 返回响应
	resp = &publish.PublishListResponse{VideoList: videoList}

	return
}
