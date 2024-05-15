package dal

import (
	"context"
	"strconv"

	"douyin/src/dal/model"
	"douyin/src/kitex_gen/comment"
	"douyin/src/pkg/snowflake"

	"gorm.io/gorm"
)

func CreateComment(ctx context.Context, comment *model.Comment) error {
	comment.ID = snowflake.GenerateID()
	return qComment.WithContext(ctx).Create(comment)
}

func DeleteComment(ctx context.Context, commentID int64) (err error) {
	_, err = qComment.WithContext(ctx).Where(qComment.ID.Eq(commentID)).Delete()
	return
}

func GetCommentByID(ctx context.Context, commentID int64) (*model.Comment, error) {
	comment, err := qComment.WithContext(ctx).
		Where(qComment.ID.Eq(commentID)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCommentNotExist
		}
		return nil, err
	}
	return comment, nil
}

func GetCommentList(ctx context.Context, videoID int64) ([]*model.Comment, error) {
	commentList, err := qComment.WithContext(ctx).Where(qComment.VideoID.Eq(videoID)).Find()
	if err != nil {
		return nil, err
	}
	return commentList, nil
}

func ToCommentResponse(ctx context.Context, userID *int64, mComment *model.Comment, user *model.User) *comment.Comment {
	return &comment.Comment{
		Id:         mComment.ID,
		User:       ToUserResponse(ctx, userID, user),
		Content:    mComment.Content,
		CreateDate: mComment.CreateTime.Format("01-02"),
	}
}

func CheckVideoExist(ctx context.Context, videoID int64) error {
	// 判断视频是否存在
	if !bloomFilter.Test([]byte(strconv.FormatInt(videoID, 10))) {
		return ErrVideoNotExist
	}

	_, err := qVideo.WithContext(ctx).
		Where(qVideo.ID.Eq(videoID)).
		Select(qVideo.ID).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrVideoNotExist
		}
		return err
	}

	return nil
}
