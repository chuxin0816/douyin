package dal

import (
	"context"

	"douyin/dal/model"
	"douyin/rpc/kitex_gen/comment"
	"strconv"

	"github.com/cloudwego/kitex/pkg/klog"
)

func CreateComment(ctx context.Context, comment *model.Comment) error {
	return qComment.WithContext(ctx).Create(comment)
}

func DeleteComment(ctx context.Context, commentID int64) (err error) {
	_, err = qComment.WithContext(ctx).Where(qComment.ID.Eq(commentID)).Delete()
	return
}

func GetCommentCount(ctx context.Context, videoID int64) (int64, error) {
	count, err := qComment.WithContext(ctx).Where(qComment.VideoID.Eq(videoID)).Count()
	if err != nil {
		return 0, err
	}
	return count, nil
}

func GetCommentByID(ctx context.Context, commentID int64) (*model.Comment, error) {
	comment, err := qComment.WithContext(ctx).
		Where(qComment.ID.Eq(commentID)).First()
	if err != nil {
		return nil, err
	}
	if comment.UserID == 0 {
		return nil, ErrCommentNotExist
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

	video, err := qVideo.WithContext(ctx).
		Where(qVideo.ID.Eq(videoID)).
		Select(qVideo.ID).First()
	if err != nil {
		klog.Error("查询视频失败, err: ", err)
		return err
	}

	if video.ID == 0 {
		return ErrVideoNotExist
	}

	return nil
}
