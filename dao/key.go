package dao

const (
	Prefix                  = "douyin:"               // 项目公共前缀
	KeyVideoCommentPF       = "video:comment:"        // Set 视频评论
	KeyVideoFavoritePF      = "video:favorite:"       // Set 视频点赞者
	KeyVideoFavoriteCountPF = "video:favorite_count:" // 视频获赞数
	KeyVideoCommentCountPF  = "video:comment_count:"  // 视频评论数
	KeyUserTotalFavoritedPF = "user:total_favorited:" // 用户总获赞数
	KeyUserFavoriteCountPF  = "user:favorite_count:"  // 用户喜欢数
	KeyUserFollowCountPF    = "user:follow_count:"    // 用户关注数
	KeyUserFollowerCountPF  = "user:follower_count:"  // 用户粉丝数
	KeyUserWorkCountPF      = "user:work_count:"      // 用户作品数
)

func getRedisKey(key string) string {
	return Prefix + key
}
