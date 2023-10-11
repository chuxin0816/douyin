package models

type ParamFeed struct {
	// 可选参数，限制返回视频的最新投稿时间戳，精确到秒，不填表示当前时间
	LatestTime *string `query:"latest_time,omitempty"`
	// 用户登录状态下设置
	Token *string `query:"token,omitempty"`
}
