package main

import (
	"context"
	"errors"
	"time"

	"douyin/src/dal"
	"douyin/src/dal/model"
	comment "douyin/src/kitex_gen/comment"
	"douyin/src/pkg/kafka"
	"douyin/src/pkg/tracing"

	"github.com/cloudwego/kitex/pkg/klog"
	"go.opentelemetry.io/otel/codes"
)

// CommentServiceImpl implements the last service interface defined in the IDL.
type CommentServiceImpl struct{}

// CommentAction implements the CommentServiceImpl interface.
func (s *CommentServiceImpl) CommentAction(ctx context.Context, req *comment.CommentActionRequest) (resp *comment.CommentActionResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "CommentAction")
	defer span.End()

	// 判断视频是否存在
	if err := CheckVideoExist(ctx, req.VideoId); err != nil {
		span.RecordError(err)

		if errors.Is(err, dal.ErrVideoNotExist) {
			span.SetStatus(codes.Error, "视频不存在")
			klog.Error("视频不存在, err: ", err)
			return nil, err
		}
		span.SetStatus(codes.Error, "查询视频失败")
		klog.Error("查询视频失败, err: ", err)
		return nil, err
	}

	mComment := &model.Comment{ID: *req.CommentId, VideoID: req.VideoId, UserID: req.UserId, Content: *req.CommentText}
	// 判断actionType
	if req.ActionType == 1 {
		// 发布评论
		mComment.CreateTime = time.Now()

		// 通过kafka异步写入数据库
		err := kafka.CreateComment(ctx, mComment)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "通过kafka异步写入数据库失败")
			klog.Error("通过kafka异步写入数据库失败, err: ", err)
			return nil, err
		}
	} else {
		// 删除评论
		mComment, err = dal.GetCommentByID(ctx, *req.CommentId)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "获取评论失败")
			klog.Error("获取评论失败", err)
			return nil, err
		}

		if mComment.UserID != req.UserId {
			span.RecordError(err)
			span.SetStatus(codes.Error, "评论作者id与当前用户id不一致")
			klog.Error("评论作者id与当前用户id不一致: ", err)
			return nil, err
		}

		// 通过kafka异步删除数据
		err := kafka.DeleteComment(ctx, *req.CommentId)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "通过kafka异步删除数据失败")
			klog.Error("通过kafka异步删除数据失败, err: ", err)
			return nil, err
		}
	}

	// 获取用户信息
	mUser, err := GetUserByID(ctx, req.UserId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取用户信息失败")
		klog.Error("获取用户信息失败, err: ", err)
		return nil, err
	}

	// 返回响应
	resp = &comment.CommentActionResponse{
		Comment: toCommentResponse(ctx, &mComment.UserID, mComment, mUser),
	}

	return
}

// CommentList implements the CommentServiceImpl interface.
func (s *CommentServiceImpl) CommentList(ctx context.Context, req *comment.CommentListRequest) (resp *comment.CommentListResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "CommentList")
	defer span.End()

	// 获取评论列表
	mCommentList, err := dal.GetCommentList(ctx, req.VideoId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取评论列表失败")
		klog.Error("获取评论列表失败, err: ", err)
		return nil, err
	}

	// 获取用户信息
	userIDs := make([]int64, len(mCommentList))
	for i, c := range mCommentList {
		userIDs[i] = c.UserID
	}
	mUsers, err := GetUserByIDs(ctx, userIDs)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取用户信息失败")
		klog.Error("获取用户信息失败, err: ", err)
		return nil, err
	}

	// 将model.Comment转换为comment.Comment
	commentList := make([]*comment.Comment, len(mCommentList))
	for i, c := range mCommentList {
		commentList[i] = toCommentResponse(ctx, req.UserId, c, mUsers[i])
	}

	// 返回响应
	resp = &comment.CommentListResponse{CommentList: commentList}

	return
}

func toCommentResponse(ctx context.Context, userID *int64, mComment *model.Comment, user *model.User) *comment.Comment {
	return &comment.Comment{
		Id:         mComment.ID,
		User:       toUserResponse(ctx, userID, user),
		Content:    mComment.Content,
		CreateDate: mComment.CreateTime.Format("01-02"),
	}
}
