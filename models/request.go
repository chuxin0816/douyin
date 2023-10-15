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

type ActionRequest struct {
	Data  *multipart.FileHeader `form:"data"`                // 视频数据
	Token string                `form:"token" vd:"len($)>0"` // 用户鉴权token
	Title string                `form:"title" vd:"len($)>0"` // 视频标题
}

type ListRequest struct {
	Token  string `query:"token" vd:"len($)>0"`     // 用户鉴权token
	UserID int64  `query:"user_id,string" vd:"$>0"` // 用户id
}
