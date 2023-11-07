package dao

import (
	"context"
	"douyin/models"
	"douyin/response"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func PublishComment(userID, commentID, videoID int64, commentText string) error {
	// 判断视频是否存在
	if !bloomFilter.Test([]byte(strconv.FormatInt(videoID, 10))) {
		return ErrVideoNotExist
	}
	video := &models.Video{ID: videoID}
	if err := db.Find(video).Error; err != nil {
		hlog.Error("mysql.PublishComment: 查询视频失败, err: ", err)
		return err
	}
	if video.AuthorID == 0 {
		return ErrVideoNotExist
	}

	// 创建评论
	comment := &models.Comment{
		ID:         commentID,
		VideoID:    videoID,
		UserID:     userID,
		Content:    commentText,
		CreateTime: time.Now(),
	}
	if err := db.Create(comment).Error; err != nil {
		hlog.Error("mysql.PublishComment: 创建评论失败, err: ", err)
		return err
	}

	// 更新video的comment_count字段
	if err := rdb.Incr(context.Background(), getRedisKey(KeyVideoCommentCountPF+strconv.FormatInt(videoID, 10))).Err(); err != nil {
		hlog.Error("redis.MessageAction: 更新video的comment_count字段失败, err: ", err)
		return err
	}
	return nil
}

func GetCommentByID(commentID int64) (*models.Comment, error) {
	comment := &models.Comment{ID: commentID}
	if err := db.Find(comment).Error; err != nil {
		hlog.Error("mysql.GetCommentUserID: 查询评论失败, err: ", err)
		return nil, err
	}
	if comment.UserID == 0 {
		return nil, ErrCommentNotExist
	}
	return comment, nil
}

func DeleteComment(commentID, videoID int64) error {
	// 判断视频是否存在
	video := &models.Video{ID: videoID}
	if err := db.Find(video).Error; err != nil {
		hlog.Error("mysql.PublishComment: 查询视频失败, err: ", err)
		return err
	}
	if video.AuthorID == 0 {
		return ErrVideoNotExist
	}

	// 删除评论
	if err := db.Delete(&models.Comment{ID: commentID}).Error; err != nil {
		hlog.Error("mysql.DeleteComment: 删除评论失败, err: ", err)
		return err
	}

	// 更新视频评论数
	if err := rdb.IncrBy(context.Background(), getRedisKey(KeyVideoCommentCountPF+strconv.FormatInt(videoID, 10)), -1).Err(); err != nil {
		hlog.Error("redis.DeleteComment: 更新视频评论数失败, err: ", err)
		return err
	}
	return nil
}

func GetCommentList(videoID int64) ([]*models.Comment, error) {
	var commentList []*models.Comment
	err := db.Where("video_id = ?", videoID).Find(&commentList).Error
	if err != nil {
		hlog.Error("mysql.GetCommentList: 查询评论列表失败, err: ", err)
		return nil, err
	}
	return commentList, nil
}

func ToCommentResponse(userID int64, comment *models.Comment, user *models.User) *response.CommentResponse {
	return &response.CommentResponse{
		ID:         comment.ID,
		User:       ToUserResponse(userID, user),
		Content:    comment.Content,
		CreateDate: comment.CreateTime.Format("01-02"),
	}
}
