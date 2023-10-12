package models

type User struct {
	ID              int64  `json:"id" gorm:"primaryKey"`               // 用户id
	Name            string `json:"name" gorm:"index"`                  // 用户名称
	Avatar          string `json:"avatar"`                             // 用户头像
	BackgroundImage string `json:"background_image"`                   // 用户个人页顶部大图
	FavoriteCount   int64  `json:"favorite_count" gorm:"default:0"`    // 喜欢数
	FollowCount     int64  `json:"follow_count" gorm:"default:0"`      // 关注总数
	FollowerCount   int64  `json:"follower_count" gorm:"default:0"`    // 粉丝总数
	WorkCount       int64  `json:"work_count" gorm:"default:0"`        // 作品数
	Password        string `json:"password"`                           // 用户密码
	Signature       string `json:"signature"`                          // 个人简介
	TotalFavorited  string `json:"total_favorited" gorm:"default:'0'"` // 获赞数量
}
