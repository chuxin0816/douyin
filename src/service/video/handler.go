package main

import (
	"context"
	"os"
	"sync"
	"time"

	"douyin/src/config"
	"douyin/src/dal"
	"douyin/src/kitex_gen/comment/commentservice"
	"douyin/src/kitex_gen/favorite/favoriteservice"
	"douyin/src/kitex_gen/user"
	"douyin/src/kitex_gen/user/userservice"
	video "douyin/src/kitex_gen/video"
	"douyin/src/pkg/oss"
	"douyin/src/pkg/tracing"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/google/uuid"
	tracing2 "github.com/kitex-contrib/obs-opentelemetry/tracing"
	consul "github.com/kitex-contrib/registry-consul"
	"go.opentelemetry.io/otel/codes"
)

const count = 30

// VideoServiceImpl implements the last service interface defined in the IDL.
type VideoServiceImpl struct{}

var (
	userClient     userservice.Client
	commentClient  commentservice.Client
	favoriteClient favoriteservice.Client
)

func init() {
	// 服务发现
	r, err := consul.NewConsulResolver(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		panic(err)
	}

	userClient, err = userservice.NewClient(
		config.Conf.OpenTelemetryConfig.UserName,
		client.WithResolver(r),
		client.WithSuite(tracing2.NewClientSuite()),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.OpenTelemetryConfig.UserName}),
		client.WithMuxConnection(2),
	)
	if err != nil {
		panic(err)
	}

	commentClient, err = commentservice.NewClient(
		config.Conf.OpenTelemetryConfig.CommentName,
		client.WithResolver(r),
		client.WithSuite(tracing2.NewClientSuite()),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.OpenTelemetryConfig.CommentName}),
		client.WithMuxConnection(2),
	)
	if err != nil {
		panic(err)
	}

	favoriteClient, err = favoriteservice.NewClient(
		config.Conf.OpenTelemetryConfig.FavoriteName,
		client.WithResolver(r),
		client.WithSuite(tracing2.NewClientSuite()),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.OpenTelemetryConfig.FavoriteName}),
		client.WithMuxConnection(2),
	)
	if err != nil {
		panic(err)
	}
}

// Feed implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) Feed(ctx context.Context, req *video.FeedRequest) (resp *video.FeedResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "Feed")
	defer span.End()

	// 参数解析
	latestTime := time.Unix(req.LatestTime, 0)
	year := latestTime.Year()
	if year < 1 || year > 9999 {
		latestTime = time.Now()
	}

	// 查询视频列表
	videoIDs, err := dal.GetFeedList(ctx, req.UserId, latestTime, count)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "查询视频列表失败")
		klog.Error("service.Feed: 查询视频列表失败, err: ", err)
		return nil, err
	}

	videoList, err := s.VideoInfoList(ctx, &video.VideoInfoListRequest{
		UserId:      req.UserId,
		VideoIdList: videoIDs,
	})

	// 计算下次请求的时间
	var nextTime *int64
	if len(videoList) > 0 {
		nextTime = new(int64)
		*nextTime = videoList[len(videoList)-1].UploadTime
	}

	// 返回响应
	resp = &video.FeedResponse{VideoList: videoList, NextTime: nextTime}
	return
}

// PublishAction implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) PublishAction(ctx context.Context, req *video.PublishActionRequest) (resp *video.PublishActionResponse, err error) {
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
	resp = &video.PublishActionResponse{}

	return
}

// PublishList implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) PublishList(ctx context.Context, req *video.PublishListRequest) (resp *video.PublishListResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "PublishList")
	defer span.End()

	// 查询视频列表
	videoIDs, err := dal.GetPublishList(ctx, req.AuthorId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "查询视频列表失败")
		klog.Error("查询视频列表失败, err: ", err)
		return nil, err
	}

	videoList, err := s.VideoInfoList(ctx, &video.VideoInfoListRequest{
		UserId:      req.UserId,
		VideoIdList: videoIDs,
	})

	// 返回响应
	resp = &video.PublishListResponse{VideoList: videoList}

	return
}

// WorkCount implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) WorkCount(ctx context.Context, userId int64) (resp int64, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "WorkCount")
	defer span.End()

	resp, err = dal.GetUserWorkCount(ctx, userId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "查询作品数量失败")
		klog.Error("查询作品数量失败, err: ", err)
		return
	}

	return
}

// AuthorId implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) AuthorId(ctx context.Context, videoId int64) (resp int64, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "AuthorId")
	defer span.End()

	resp, err = dal.GetAuthorID(ctx, videoId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "查询作者ID失败")
		klog.Error("查询作者ID失败, err: ", err)
		return
	}

	return
}

// VideoExist implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) VideoExist(ctx context.Context, videoId int64) (resp bool, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "VideoExist")
	defer span.End()

	resp, err = dal.CheckVideoExist(ctx, videoId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "查询视频是否存在失败")
		klog.Error("查询视频是否存在失败, err: ", err)
		return
	}

	return
}

// VideoInfo implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) VideoInfo(ctx context.Context, req *video.VideoInfoRequest) (resp *video.Video, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "VideoInfo")
	defer span.End()

	mVideo, err := dal.GetVideoByID(ctx, req.VideoId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "查询视频信息失败")
		klog.Error("查询视频信息失败, err: ", err)
		return
	}

	resp = &video.Video{
		Id:         mVideo.ID,
		PlayUrl:    mVideo.PlayURL,
		CoverUrl:   mVideo.CoverURL,
		IsFavorite: false,
		Title:      mVideo.Title,
	}

	var wg sync.WaitGroup
	var wgErr error
	wg.Add(3)
	go func() {
		defer wg.Done()
		user, err := userClient.UserInfo(ctx, &user.UserInfoRequest{
			UserId:   req.UserId,
			AuthorId: mVideo.AuthorID,
		})
		if err != nil {
			wgErr = err
			return
		}
		resp.Author = user.User
	}()
	go func() {
		defer wg.Done()
		cnt, err := commentClient.CommentCnt(ctx, mVideo.ID)
		if err != nil {
			wgErr = err
			return
		}
		resp.CommentCount = cnt
	}()
	go func() {
		defer wg.Done()
		cnt, err := favoriteClient.FavoriteCnt(ctx, mVideo.ID)
		if err != nil {
			wgErr = err
			return
		}
		resp.FavoriteCount = cnt
	}()
	wg.Wait()
	if wgErr != nil {
		span.RecordError(wgErr)
		span.SetStatus(codes.Error, "查询视频信息失败")
		klog.Error("查询视频信息失败, err: ", wgErr)
		return nil, wgErr
	}

	// 未登录直接返回
	if req.UserId == nil {
		return
	}

	// 查询缓存判断是否点赞
	exist, err := favoriteClient.FavoriteExist(ctx, *req.UserId, mVideo.ID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "查询视频是否点赞失败")
		klog.Error("查询视频是否点赞失败, err: ", err)
		return
	}
	resp.IsFavorite = exist

	return
}

// VideoInfoList implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) VideoInfoList(ctx context.Context, req *video.VideoInfoListRequest) (resp []*video.Video, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "VideoInfoList")
	defer span.End()

	resp = make([]*video.Video, len(req.VideoIdList))
	for i, videoId := range req.VideoIdList {
		video, err := s.VideoInfo(ctx, &video.VideoInfoRequest{
			UserId:  req.UserId,
			VideoId: videoId,
		})
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "查询视频信息失败")
			klog.Error("查询视频信息失败, err: ", err)
			return nil, err
		}

		resp[i] = video
	}

	return
}

// PublishIDList implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) PublishIDList(ctx context.Context, userId int64) (resp []int64, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "PublishIDList")
	defer span.End()

	resp, err = dal.GetPublishList(ctx, userId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "查询视频ID列表失败")
		klog.Error("查询视频ID列表失败, err: ", err)
		return
	}

	return
}
