package response

type FavoriteActionResponse struct {
	*Response
}

type FavoriteListResponse struct {
	*Response
	VideoList []*VideoResponse `json:"video_list"`
}
