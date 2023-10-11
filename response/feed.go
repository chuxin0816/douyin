package response

import "github.com/chuxin0816/Scaffold/models"

type FeedResponse struct {
	*Response
	NextTime  int64          `json:"next_time"`  // 本次返回的视频中，发布最早的时间，作为下次请求时的latest_time
	VideoList []models.Video `json:"video_list"` // 视频列表
}
