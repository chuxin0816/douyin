package dal

import (
	"context"

	"douyin/dal/model"
	"douyin/rpc/kitex_gen/comment"
	"strconv"

	"github.com/cloudwego/kitex/pkg/klog"
)

func CreateComment(ctx context.Context, comment *model.Comment) error {
	return qComment.WithContext(context.Background()).Create(comment)
}

func DeleteComment(ctx context.Context, commentID int64) (err error) {
	_, err = qComment.WithContext(ctx).Where(qComment.ID.Eq(commentID)).Delete()
	return
}

func GetCommentByID(commentID int64) (*model.Comment, error) {
	comment, err := qComment.WithContext(context.Background()).
		Where(qComment.ID.Eq(commentID)).First()
	if err != nil {
		klog.Error("查询评论失败, err: ", err)
		return nil, err
	}
	if comment.UserID == 0 {
		return nil, ErrCommentNotExist
	}
	return comment, nil
}

func GetCommentList(videoID int64) ([]*model.Comment, error) {
	commentList, err := qComment.WithContext(context.Background()).Where(qComment.VideoID.Eq(videoID)).Find()
	if err != nil {
		klog.Error("查询评论列表失败, err: ", err)
		return nil, err
	}
	return commentList, nil
}

func ToCommentResponse(userID *int64, mComment *model.Comment, user *model.User) *comment.Comment {
	return &comment.Comment{
		Id:         mComment.ID,
		User:       ToUserResponse(userID, user),
		Content:    mComment.Content,
		CreateDate: mComment.CreateTime.Format("01-02"),
	}
}

func CheckVideoExist(videoID int64) error {
	// 判断视频是否存在
	if !bloomFilter.Test([]byte(strconv.FormatInt(videoID, 10))) {
		return ErrVideoNotExist
	}

	video, err := qVideo.WithContext(context.Background()).
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
