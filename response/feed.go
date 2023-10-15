package response

type FeedResponse struct {
	*Response
	VideoList []*VideoResponse `json:"video_list"` // 视频列表
	NextTime  *int64           `json:"next_time"`  // 本次返回的视频中，发布最早的时间，作为下次请求时的latest_time
}
