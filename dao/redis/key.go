package redis

const (
	Prefix           = "douyin:"
	KeyVideoFavorite = "video:favorite"
	KeyVideoLiker    = "video:liker"
	KeyVideoComment  = "video:comment"
)

func getRedisKey(key string) string {
	return Prefix + key
}
