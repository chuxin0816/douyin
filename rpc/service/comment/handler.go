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
	if err := dal.CheckVideoExist(ctx, req.VideoId); err != nil {
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
		go func() {
			key := dal.GetRedisKey(dal.KeyVideoCommentCountPF + strconv.FormatInt(req.VideoId, 10))
			// 检查缓存是否存在
			if exist, err := dal.RDB.Exists(ctx, key).Result(); err != nil {
				klog.Error("查询缓存失败, err: ", err)
				return
			} else if exist == 0 {
				// 缓存不存在，查询数据库写入缓存
				cnt, err := dal.GetCommentCount(ctx, req.VideoId)
				if err != nil {
					klog.Error("查询评论数量失败, err: ", err)
					return
				}
				if err := dal.RDB.Set(ctx, key, cnt, 0).Err(); err != nil {
					klog.Error("写入缓存失败, err: ", err)
					return
				}
			}
			if err := dal.RDB.Incr(ctx, key).Err(); err != nil {
				klog.Error("更新video的comment_count字段失败, err: ", err)
				return
			}

			// 写入待同步切片
			dal.CacheVideoID.Store(req.VideoId, struct{}{})
		}()
	} else {
		// 删除评论
		mComment, err = dal.GetCommentByID(ctx, *req.CommentId)
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
		go func() {
			key := dal.GetRedisKey(dal.KeyVideoCommentCountPF + strconv.FormatInt(req.VideoId, 10))
			// 检查缓存是否存在
			if exist, err := dal.RDB.Exists(ctx, key).Result(); err != nil {
				klog.Error("查询缓存失败, err: ", err)
				return
			} else if exist == 0 {
				// 缓存不存在，查询数据库写入缓存
				cnt, err := dal.GetCommentCount(ctx, req.VideoId)
				if err != nil {
					klog.Error("查询评论数量失败, err: ", err)
					return
				}
				if err := dal.RDB.Set(ctx, key, cnt, 0).Err(); err != nil {
					klog.Error("写入缓存失败, err: ", err)
					return
				}
			}
			if err := dal.RDB.IncrBy(ctx, key, -1).Err(); err != nil {
				klog.Error("更新视频评论数失败, err: ", err)
			}
		}()
	}

	// 获取用户信息
	mUser, err := dal.GetUserByID(ctx, req.UserId)
	if err != nil {
		klog.Error("获取用户信息失败, err: ", err)
		return nil, err
	}

	// 返回响应
	resp = &comment.CommentActionResponse{
		Comment: dal.ToCommentResponse(ctx, &mComment.UserID, mComment, mUser),
	}

	return
}

// CommentList implements the CommentServiceImpl interface.
func (s *CommentServiceImpl) CommentList(ctx context.Context, req *comment.CommentListRequest) (resp *comment.CommentListResponse, err error) {
	// 获取评论列表
	mCommentList, err := dal.GetCommentList(ctx, req.VideoId)
	if err != nil {
		klog.Error("获取评论列表失败, err: ", err)
		return nil, err
	}

	// 获取用户信息
	userIDs := make([]int64, len(mCommentList))
	for i, c := range mCommentList {
		userIDs[i] = c.UserID
	}
	mUsers, err := dal.GetUserByIDs(ctx, userIDs)
	if err != nil {
		klog.Error("获取用户信息失败, err: ", err)
		return nil, err
	}

	// 将用户信息与评论列表进行关联
	commentList := make([]*comment.Comment, len(mCommentList))
	for i, c := range mCommentList {
		commentList[i] = dal.ToCommentResponse(ctx, req.UserId, c, mUsers[i])
	}

	// 返回响应
	resp = &comment.CommentListResponse{CommentList: commentList}

	return
}
