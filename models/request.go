package models

import "mime/multipart"

type FeedRequest struct {
	LatestTime int64  `query:"latest_time,string"` // 可选参数，限制返回视频的最新投稿时间戳，精确到秒，不填表示当前时间
	Token      string `query:"token"`              // 用户登录状态下设置
}

type UserRequest struct {
	Username string `query:"username" vd:"0<len($)&&len($)<33"` // 注册用户名，最长32个字符
	Password string `query:"password" vd:"5<len($)&&len($)<33"` // 密码，最长32个字符
}

type UserInfoRequest struct {
	UserID int64  `query:"user_id,string" vd:"$>0"` // 用户id
	Token  string `query:"token" vd:"len($)>0"`     // 用户登录状态下设置
}

type PublishActionRequest struct {
	Data  *multipart.FileHeader `form:"data"`                // 视频数据
	Token string                `form:"token" vd:"len($)>0"` // 用户鉴权token
	Title string                `form:"title" vd:"len($)>0"` // 视频标题
}

type PublishListRequest struct {
	Token  string `query:"token" vd:"len($)>0"`     // 用户鉴权token
	UserID int64  `query:"user_id,string" vd:"$>0"` // 用户id
}

type FavoriteActionRequest struct {
	Token      string `query:"token" vd:"len($)>0"`                // 用户鉴权token
	VideoID    int64  `query:"video_id,string" vd:"$>0"`           // 视频id
	ActionType int    `query:"action_type,string" vd:"$==1||$==2"` // 1-点赞，2-取消点赞
}

type FavoriteListRequest struct {
	UserID int64  `query:"user_id,string" vd:"$>0"` // 用户id
	Token  string `query:"token" vd:"len($)>0"`     // 用户鉴权token
}

type CommentActionRequest struct {
	Token       string `query:"token" vd:"len($)>0"`                                              // 用户鉴权token
	VideoID     int64  `query:"video_id,string" vd:"$>0"`                                         // 视频id
	ActionType  int64  `query:"action_type,string" vd:"$==1||$==2"`                               // 1-发布评论，2-删除评论
	CommentID   int64  `query:"comment_id,string" vd:"((ActionType)$==2&&$>0)||(ActionType)$==1"` // 要删除的评论id，在action_type=2的时候使用
	CommentText string `query:"comment_text" vd:"((ActionType)$==1&&len($)>0)||(ActionType)$==2"` // 用户填写的评论内容，在action_type=1的时候使用
}

type CommentListRequest struct {
	Token   string `query:"token" vd:"len($)>0"`      // 用户鉴权token
	VideoID int64  `query:"video_id,string" vd:"$>0"` // 视频id
}
