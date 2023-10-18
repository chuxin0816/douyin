package service

import (
	"douyin/dao/mysql"
	"douyin/models"
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
		err = mysql.PublishComment(userID, videoID, commentText)
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
		err = mysql.DeleteComment(commentID)
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
		Comment: &response.CommentResponse{
			ID:         commentID,
			User:       *response.ToUserResponse(user),
			Content:    commentText,
			CreateDate: comment.CreateTime.Format("01-02"),
		},
	}, nil
}
