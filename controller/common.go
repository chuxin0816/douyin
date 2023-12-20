package controller

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func Success(ctx *app.RequestContext, data any) {
	ctx.JSON(consts.StatusOK, data)
}

func Error(ctx *app.RequestContext, code ResCode) {
	ctx.JSON(consts.StatusOK, &Response{StatusCode: code, StatusMsg: code.Msg()})
}

func ErrorWithMsg(ctx *app.RequestContext, code ResCode, msg string) {
	ctx.JSON(consts.StatusOK, &Response{StatusCode: code, StatusMsg: msg})
}

type Response struct {
	StatusCode ResCode `json:"status_code"` // 状态码，0-成功，其他值-失败
	StatusMsg  string  `json:"status_msg"`  // 返回状态描述
}

type UserResponse struct {
	ID              int64  `json:"id"`               // 用户id
	Name            string `json:"name"`             // 用户名称
	Avatar          string `json:"avatar"`           // 用户头像
	BackgroundImage string `json:"background_image"` // 用户个人页顶部大图
	FavoriteCount   int64  `json:"favorite_count"`   // 喜欢数
	FollowCount     int64  `json:"follow_count"`     // 关注总数
	FollowerCount   int64  `json:"follower_count"`   // 粉丝总数
	WorkCount       int64  `json:"work_count"`       // 作品数
	IsFollow        bool   `json:"is_follow"`        // true-已关注，false-未关注
	Signature       string `json:"signature"`        // 个人简介
	TotalFavorited  string `json:"total_favorited"`  // 获赞数量
}

type VideoResponse struct {
	ID            int64         `json:"id"`             // 视频唯一标识
	Author        *UserResponse `json:"author"`         // 视频作者信息
	CommentCount  int64         `json:"comment_count"`  // 视频的评论总数
	PlayURL       string        `json:"play_url"`       // 视频播放地址
	CoverURL      string        `json:"cover_url"`      // 视频封面地址
	FavoriteCount int64         `json:"favorite_count"` // 视频的点赞总数
	IsFavorite    bool          `json:"is_favorite"`    // true-已点赞，false-未点赞
	Title         string        `json:"title"`          // 视频标题
}

type CommentResponse struct {
	ID         int64         `json:"id"`          // 评论id
	User       *UserResponse `json:"user"`        // 评论用户信息
	Content    string        `json:"content"`     // 评论内容
	CreateDate string        `json:"create_date"` // 评论发布日期，格式 mm-dd
}

type MessageResponse struct {
	ID         int64  `json:"id"`           // 消息id
	ToUserID   int64  `json:"to_user_id"`   // 消息接收者id
	FromUserID int64  `json:"from_user_id"` // 消息发送者id
	Content    string `json:"content"`      // 消息内容
	CreateTime int64  `json:"create_time"`  // 消息发送时间 yyyy-MM-dd HH:MM:ss
}
