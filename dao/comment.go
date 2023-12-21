package dao

import (
	"context"

	"douyin/dao/model"
	"douyin/rpc/kitex_gen/comment"
	"strconv"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
)

func PublishComment(userID, commentID, videoID int64, commentText string) error {
	// 判断视频是否存在
	if !bloomFilter.Test([]byte(strconv.FormatInt(videoID, 10))) {
		return ErrVideoNotExist
	}
	video, err := qVideo.WithContext(context.Background()).
		Where(qVideo.ID.Eq(videoID)).
		Select(qVideo.AuthorID).First()
	if err != nil {
		klog.Error("mysql.PublishComment: 查询视频失败, err: ", err)
		return err
	}
	if video.AuthorID == 0 {
		return ErrVideoNotExist
	}

	// 创建评论
	comment := &model.Comment{
		ID:         commentID,
		VideoID:    videoID,
		UserID:     userID,
		Content:    commentText,
		CreateTime: time.Now(),
	}

	if err := qComment.WithContext(context.Background()).Create(comment); err != nil {
		klog.Error("mysql.PublishComment: 创建评论失败, err: ", err)
		return err
	}

	// 更新video的comment_count字段
	if err := rdb.Incr(context.Background(), getRedisKey(KeyVideoCommentCountPF+strconv.FormatInt(videoID, 10))).Err(); err != nil {
		klog.Error("redis.PublishMessage: 更新video的comment_count字段失败, err: ", err)
		return err
	}

	// 写入待同步切片
	lock.Lock()
	cacheVideoIDs = append(cacheVideoIDs, videoID)
	lock.Unlock()

	return nil
}

func GetCommentByID(commentID int64) (*model.Comment, error) {
	comment, err := qComment.WithContext(context.Background()).
		Where(qComment.ID.Eq(commentID)).First()
	if err != nil {
		klog.Error("mysql.GetCommentUserID: 查询评论失败, err: ", err)
		return nil, err
	}
	if comment.UserID == 0 {
		return nil, ErrCommentNotExist
	}
	return comment, nil
}

func DeleteComment(commentID, videoID int64) error {
	// 判断视频是否存在
	video, err := qVideo.WithContext(context.Background()).
		Where(qVideo.ID.Eq(videoID)).
		Select(qVideo.ID).First()
	if err != nil {
		klog.Error("mysql.PublishComment: 查询视频失败, err: ", err)
		return err
	}
	if video.ID == 0 {
		return ErrVideoNotExist
	}

	// 删除评论
	if _, err := qComment.WithContext(context.Background()).Where(qComment.ID.Eq(commentID)).Delete(); err != nil {
		klog.Error("mysql.DeleteComment: 删除评论失败, err: ", err)
		return err
	}

	// 更新视频评论数
	if err := rdb.IncrBy(context.Background(), getRedisKey(KeyVideoCommentCountPF+strconv.FormatInt(videoID, 10)), -1).Err(); err != nil {
		klog.Error("redis.DeleteComment: 更新视频评论数失败, err: ", err)
		return err
	}
	return nil
}

func GetCommentList(videoID int64) ([]*model.Comment, error) {
	commentList, err := qComment.WithContext(context.Background()).Where(qComment.VideoID.Eq(videoID)).Find()
	if err != nil {
		klog.Error("mysql.GetCommentList: 查询评论列表失败, err: ", err)
		return nil, err
	}
	return commentList, nil
}

func ToCommentResponse(userID int64, mComment *model.Comment, user *model.User) *comment.Comment {
	return &comment.Comment{
		Id:         mComment.ID,
		User:       ToUserResponse(userID, user),
		Content:    mComment.Content,
		CreateDate: mComment.CreateTime.Format("01-02"),
	}
}
