package dao

const (
	Prefix           = "douyin:"
	KeyVideoFavorite = "video:favorite"
	KeyVideoLikerPF    = "video:liker:"
	KeyVideoComment  = "video:comment"
)

func getRedisKey(key string) string {
	return Prefix + key
}
