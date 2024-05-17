package dal

import "strings"

const (
	Prefix                  = "douyin:"               // 项目公共前缀
	KeyVideoInfoPF          = "video:info:"           // 视频基础信息
	KeyVideoFavoriteCountPF = "video:favorite_count:" // 视频获赞数
	KeyVideoCommentCountPF  = "video:comment_count:"  // 视频评论数
	KeyUserFavoritePF       = "user:favorite:"        // Set 用户喜欢的视频
	KeyUserFollowPF         = "user:follow:"          // Set 用户关注列表
	KeyUserFollowerPF       = "user:follower:"        // Set 用户粉丝列表(50条左右)
	KeyUserFriendPF         = "user:friend:"          // Set 用户好友列表
	KeyUserInfoPF           = "user:info:"            // 用户基础信息
	KeyUserTotalFavoritedPF = "user:total_favorited:" // 用户总获赞数
	KeyUserFavoriteCountPF  = "user:favorite_count:"  // 用户喜欢数
	KeyUserFollowCountPF    = "user:follow_count:"    // 用户关注数
	KeyUserFollowerCountPF  = "user:follower_count:"  // 用户粉丝数
	KeyUserWorkCountPF      = "user:work_count:"      // 用户作品数
)

func GetRedisKey(keys ...string) string {
	var builder strings.Builder
	builder.WriteString(Prefix)
	for _, key := range keys {
		builder.WriteString(key)
	}
	return builder.String()
}
