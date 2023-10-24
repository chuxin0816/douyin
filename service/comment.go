package service

import (
	"douyin/dao/mysql"
	"douyin/models"
	"douyin/pkg/snowflake"
	"douyin/response"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func CommentAction(userID, actionType, videoID, commentID int64, commentText string) (*response.CommentActionResponse, error) {
	comment := &models.Comment{ID: commentID, VideoID: videoID, UserID: userID, Content: commentText, CreateTime: time.Now()}
	var err error
	// 判断actionType
	if actionType == 1 {
		// 发布评论
		commentID = snowflake.GenerateID()
		err = mysql.PublishComment(userID, commentID, videoID, commentText)
		if err != nil {
			hlog.Error("service.CommentAction: 发布评论失败, err: ", err)
			return nil, err
		}
	} else {
		// 删除评论
		comment, err = mysql.GetCommentByID(commentID)
		if err != nil {
			hlog.Error("service.CommentAction: 获取评论作者id失败, err: ", err)
			return nil, err
		}
		if comment.UserID != userID {
			hlog.Error("service.CommentAction: 评论作者id与当前用户id不一致")
			return nil, err
		}
		err = mysql.DeleteComment(commentID, videoID)
		if err != nil {
			hlog.Error("service.CommentAction: 删除评论失败, err: ", err)
			return nil, err
		}
	}

	// 获取用户信息
	user, err := mysql.GetUserByID(userID, comment.UserID)
	if err != nil {
		hlog.Error("service.CommentAction: 获取用户信息失败, err: ", err)
		return nil, err
	}

	// 返回响应
	return &response.CommentActionResponse{
		Response: &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
		Comment:  mysql.ToCommentResponse(userID, comment, user),
	}, nil
}

func CommentList(userID, videoID int64) (*response.CommentListResponse, error) {
	// 获取评论列表
	dCommentList, err := mysql.GetCommentList(videoID)
	if err != nil {
		hlog.Error("service.CommentList: 获取评论列表失败, err: ", err)
		return nil, err
	}

	// 获取用户信息
	userIDs := make([]int64, 0, len(dCommentList))
	for _, comment := range dCommentList {
		userIDs = append(userIDs, comment.UserID)
	}
	users, err := mysql.GetUserByIDs(userID, userIDs)
	if err != nil {
		hlog.Error("service.CommentList: 获取用户信息失败, err: ", err)
		return nil, err
	}

	// 将用户信息与评论列表进行关联
	commentList := make([]*response.CommentResponse, 0, len(dCommentList))
	for idx, comment := range dCommentList {
		commentList = append(commentList, mysql.ToCommentResponse(userID, comment, users[idx]))
	}

	// 返回响应
	return &response.CommentListResponse{
		Response:    &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
		CommentList: commentList,
	}, nil
}
