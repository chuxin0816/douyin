package main

import (
	"context"
	"douyin/dal"
	"douyin/dal/model"
	"douyin/pkg/snowflake"
	comment "douyin/rpc/kitex_gen/comment"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

// CommentServiceImpl implements the last service interface defined in the IDL.
type CommentServiceImpl struct{}

// CommentAction implements the CommentServiceImpl interface.
func (s *CommentServiceImpl) CommentAction(ctx context.Context, req *comment.CommentActionRequest) (resp *comment.CommentActionResponse, err error) {
	mComment := &model.Comment{ID: *req.CommentId, VideoID: req.VideoId, UserID: req.UserId, Content: *req.CommentText}
	// 判断actionType
	if req.ActionType == 1 {
		// 发布评论
		mComment.ID = snowflake.GenerateID()
		mComment.CreateTime = time.Now()
		if err := dal.PublishComment(req.UserId, mComment.ID, req.VideoId, *req.CommentText); err != nil {
			klog.Error("发布评论失败, err: ", err)
			return nil, err
		}
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

		if err := dal.DeleteComment(*req.CommentId, req.VideoId); err != nil {
			klog.Error("删除评论失败", err)
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
		Comment: dal.ToCommentResponse(mComment.UserID, mComment, mUser),
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
