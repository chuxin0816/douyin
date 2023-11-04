package dao

import (
	"douyin/models"
	"douyin/response"
	"errors"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/gorm"
)

var (
	ErrCommentNotExist = errors.New("comment not exist")
	ErrVideoNotExist   = errors.New("video not exist")
)

func PublishComment(userID, commentID, videoID int64, commentText string) error {
	// 判断视频是否存在
	video := &models.Video{ID: videoID}
	err := db.Find(video).Error
	if err != nil {
		hlog.Error("mysql.PublishComment: 查询视频失败, err: ", err)
		return err
	}
	if video.AuthorID == 0 {
		hlog.Error("mysql.PublishComment: 视频不存在")
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
	err = db.Create(comment).Error
	if err != nil {
		hlog.Error("mysql.PublishComment: 创建评论失败, err: ", err)
		return err
	}

	// 更新视频评论数
	err = db.Model(&models.Video{}).Where("id = ?", videoID).Update("comment_count", gorm.Expr("comment_count + ?", 1)).Error
	if err != nil {
		hlog.Error("mysql.PublishComment: 更新视频评论数失败, err: ", err)
		return err
	}
	return nil
}

func GetCommentByID(commentID int64) (*models.Comment, error) {
	comment := &models.Comment{ID: commentID}
	err := db.Find(comment).Error
	if err != nil {
		hlog.Error("mysql.GetCommentUserID: 查询评论失败, err: ", err)
		return nil, err
	}
	if comment.UserID == 0 {
		hlog.Error("mysql.GetCommentUserID: 评论不存在")
		return nil, ErrCommentNotExist
	}
	return comment, nil
}

func DeleteComment(commentID, videoID int64) error {
	// 判断视频是否存在
	video := &models.Video{ID: videoID}
	err := db.Find(video).Error
	if err != nil {
		hlog.Error("mysql.PublishComment: 查询视频失败, err: ", err)
		return err
	}
	if video.AuthorID == 0 {
		hlog.Error("mysql.PublishComment: 视频不存在")
		return ErrVideoNotExist
	}
	// 删除评论
	err = db.Delete(&models.Comment{ID: commentID}).Error
	if err != nil {
		hlog.Error("mysql.DeleteComment: 删除评论失败, err: ", err)
		return err
	}
	// 更新视频评论数
	err = db.Model(&models.Video{}).Where("id = ?", videoID).Update("comment_count", gorm.Expr("comment_count - ?", 1)).Error
	if err != nil {
		hlog.Error("mysql.DeleteComment: 更新视频评论数失败, err: ", err)
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
