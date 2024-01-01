package main

import (
	"context"
	"douyin/dal"
	"douyin/dal/model"
	"douyin/pkg/kafka"
	"douyin/pkg/snowflake"
	comment "douyin/rpc/kitex_gen/comment"
	"errors"
	"strconv"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

// CommentServiceImpl implements the last service interface defined in the IDL.
type CommentServiceImpl struct{}

// CommentAction implements the CommentServiceImpl interface.
func (s *CommentServiceImpl) CommentAction(ctx context.Context, req *comment.CommentActionRequest) (resp *comment.CommentActionResponse, err error) {
	// 判断视频是否存在
	if err := dal.CheckVideoExist(req.VideoId); err != nil {
		if errors.Is(err, dal.ErrVideoNotExist) {
			klog.Error("视频不存在, err: ", err)
			return nil, err
		}
		klog.Error("查询视频失败, err: ", err)
		return nil, err
	}

	mComment := &model.Comment{ID: *req.CommentId, VideoID: req.VideoId, UserID: req.UserId, Content: *req.CommentText}
	// 判断actionType
	if req.ActionType == 1 {
		// 发布评论
		mComment.ID = snowflake.GenerateID()
		mComment.CreateTime = time.Now()

		// 通过kafka异步写入数据库
		err := kafka.CreateComment(ctx, mComment)
		if err != nil {
			klog.Error("通过kafka异步写入数据库失败, err: ", err)
			return nil, err
		}

		// 更新video的comment_count字段
		if err := dal.RDB.Incr(context.Background(), dal.GetRedisKey(dal.KeyVideoCommentCountPF+strconv.FormatInt(req.VideoId, 10))).Err(); err != nil {
			klog.Error("更新video的comment_count字段失败, err: ", err)
		}

		// 写入待同步切片
		dal.CacheVideoID.Store(req.VideoId, struct{}{})
	} else {
		// 删除评论
		mComment, err = dal.GetCommentByID(*req.CommentId)
		if err != nil {
			klog.Error("获取评论失败", err)
			return nil, err
		}

		if mComment.UserID != req.UserId {
			klog.Error("评论作者id与当前用户id不一致: ", err)
			return nil, err
		}

		// 通过kafka异步删除数据
		err := kafka.DeleteComment(ctx, *req.CommentId)
		if err != nil {
			klog.Error("通过kafka异步删除数据失败, err: ", err)
			return nil, err
		}

		// 更新视频评论数
		if err := dal.RDB.IncrBy(context.Background(), dal.GetRedisKey(dal.KeyVideoCommentCountPF+strconv.FormatInt(req.VideoId, 10)), -1).Err(); err != nil {
			klog.Error("更新视频评论数失败, err: ", err)
			return nil, err
		}
	}

	// 获取用户信息
	mUser, err := dal.GetUserByID(req.UserId)
	if err != nil {
		klog.Error("获取用户信息失败, err: ", err)
		return nil, err
	}

	// 返回响应
	resp = &comment.CommentActionResponse{
		Comment: dal.ToCommentResponse(&mComment.UserID, mComment, mUser),
	}

	return
}

// CommentList implements the CommentServiceImpl interface.
func (s *CommentServiceImpl) CommentList(ctx context.Context, req *comment.CommentListRequest) (resp *comment.CommentListResponse, err error) {
	// 获取评论列表
	mCommentList, err := dal.GetCommentList(req.VideoId)
	if err != nil {
		klog.Error("获取评论列表失败, err: ", err)
		return nil, err
	}

	// 获取用户信息
	userIDs := make([]int64, len(mCommentList))
	for i, c := range mCommentList {
		userIDs[i] = c.UserID
	}
	mUsers, err := dal.GetUserByIDs(userIDs)
	if err != nil {
		klog.Error("获取用户信息失败, err: ", err)
		return nil, err
	}

	// 将用户信息与评论列表进行关联
	commentList := make([]*comment.Comment, len(mCommentList))
	for i, c := range mCommentList {
		commentList[i] = dal.ToCommentResponse(req.UserId, c, mUsers[i])
	}

	// 返回响应
	resp = &comment.CommentListResponse{CommentList: commentList}

	return
}
