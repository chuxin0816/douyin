package response

import (
	"douyin/models"
)

type CommentResponse struct {
	ID         int64         `json:"id"`          // 评论id
	User       *UserResponse `json:"user"`        // 评论用户信息
	Content    string        `json:"content"`     // 评论内容
	CreateDate string        `json:"create_date"` // 评论发布日期，格式 mm-dd
}

type CommentActionResponse struct {
	*Response
	Comment *CommentResponse `json:"comment,omitempty"`
}

type CommentListResponse struct {
	*Response
	CommentList []*CommentResponse `json:"comment_list"`
}

func ToCommentResponse(comment *models.Comment, user *models.User) *CommentResponse {
	return &CommentResponse{
		ID:         comment.ID,
		User:       ToUserResponse(user),
		Content:    comment.Content,
		CreateDate: comment.CreateTime.Format("01-02"),
	}
}
