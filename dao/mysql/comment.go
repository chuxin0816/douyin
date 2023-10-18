package mysql

import (
	"douyin/models"
	"douyin/pkg/snowflake"
	"errors"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

var (
	ErrCommentNotExist = errors.New("comment not exist")
)

func PublishComment(userID, videoID int64, commentText string) error {
	comment := &models.Comment{
		ID:         snowflake.GenerateID(),
		VideoID:    videoID,
		UserID:     userID,
		Content:    commentText,
		CreateTime: time.Now(),
	}
	err := db.Create(comment).Error
	if err != nil {
		hlog.Error("mysql.PublishComment: 创建评论失败, err: ", err)
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

func DeleteComment(commentID int64) error {
	err := db.Delete(&models.Comment{ID: commentID}).Error
	if err != nil {
		hlog.Error("mysql.DeleteComment: 删除评论失败, err: ", err)
		return err
	}
	return nil
}
