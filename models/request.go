package models

type FeedRequest struct {
	LatestTime string `query:"latest_time"` // 可选参数，限制返回视频的最新投稿时间戳，精确到秒，不填表示当前时间
	Token      string `query:"token"`       // 用户登录状态下设置
}

type UserRequest struct {
	Username string `query:"username" vd:"0<len($)&&len($)<33"` // 注册用户名，最长32个字符
	Password string `query:"password" vd:"0<len($)&&len($)<33"` // 密码，最长32个字符
}
